/* SPDX-License-Identifier: BSD-3-Clause */
/* Copyright 2015-2024, Intel Corporation */

/*
 * palloc.c -- implementation of pmalloc POSIX-like API
 *
 * This is the front-end part of the persistent memory allocator. It uses both
 * transient and persistent representation of the heap to provide memory blocks
 * in a reasonable time and with an acceptable common-case fragmentation.
 *
 * Lock ordering in the entirety of the allocator is simple, but might be hard
 * to follow at times because locks are, by necessity, externalized.
 * There are two sets of locks that need to be taken into account:
 *	- runtime state locks, represented by buckets.
 *	- persistent state locks, represented by memory block mutexes.
 *
 * To properly use them, follow these rules:
 *	- When nesting, always lock runtime state first.
 *	Doing the reverse might cause deadlocks in other parts of the code.
 *
 *	- When introducing functions that would require runtime state locks,
 *	always try to move the lock acquiring to the upper most layer. This
 *	usually means that the functions will simply take "struct bucket" as
 *	their argument. By doing so most of the locking can happen in
 *	the frontend part of the allocator and it's easier to follow the first
 *	rule because all functions in the backend can safely use the persistent
 *	state locks - the runtime lock, if it is needed, will be already taken
 *	by the upper layer.
 *
 * General lock ordering:
 *	1. arenas.lock
 *	2. buckets (sorted by ID)
 *	3. memory blocks (sorted by lock address)
 */

#include "bucket.h"
#include "valgrind_internal.h"
#include "heap_layout.h"
#include "heap.h"
#include "alloc_class.h"
#include "out.h"
#include "sys_util.h"
#include "palloc.h"
#include "ravl.h"
#include "vec.h"

struct dav_action_internal {
	/* type of operation (alloc/free vs set) */
	enum dav_action_type type;

	uint32_t padding;

	/*
	 * Action-specific lock that needs to be taken for the duration of
	 * an action.
	 */
	pthread_mutex_t *lock;

	/* action-specific data */
	union {
		/* valid only when type == DAV_ACTION_TYPE_HEAP */
		struct {
			uint64_t offset;
			uint64_t usable_size;
			enum memblock_state new_state;
			struct memory_block m;
			struct memory_block_reserved *mresv;
		};

		/* valid only when type == DAV_ACTION_TYPE_MEM */
		struct {
			uint64_t *ptr;
			uint64_t value;
		};

		/* padding, not used */
		uint64_t data2[14];
	};
};
D_CASSERT(offsetof(struct dav_action_internal, data2) == offsetof(struct dav_action, data2),
	  "struct dav_action misaligned!");

/*
 * palloc_set_value -- creates a new set memory action
 */
void
palloc_set_value(struct palloc_heap *heap, struct dav_action *act,
	uint64_t *ptr, uint64_t value)
{
	/* suppress unused-parameter errors */
	SUPPRESS_UNUSED(heap);

	act->type = DAV_ACTION_TYPE_MEM;

	struct dav_action_internal *actp = (struct dav_action_internal *)act;

	actp->ptr = ptr;
	actp->value = value;
	actp->lock = NULL;
}

static void *
zone_get_base_address(struct palloc_heap *heap, void *ptr)
{
	uint64_t off = HEAP_PTR_TO_OFF(heap, ptr);
	uint32_t zid = heap_off2mbid(heap, off);

	if (zid)
		return ZID_TO_ZONE(&heap->layout_info, zid);

	return heap->layout_info.zone0;
}

/*
 * alloc_prep_block -- (internal) prepares a memory block for allocation
 *
 * Once the block is fully reserved and it's guaranteed that no one else will
 * be able to write to this memory region it is safe to write the allocation
 * header and call the object construction function.
 *
 * Because the memory block at this stage is only reserved in transient state
 * there's no need to worry about fail-safety of this method because in case
 * of a crash the memory will be back in the free blocks collection.
 */
static int
alloc_prep_block(struct palloc_heap *heap, const struct memory_block *m,
	palloc_constr constructor, void *arg,
	uint64_t extra_field, uint16_t object_flags,
	struct dav_action_internal *out)
{
	void *uptr = m->m_ops->get_user_data(m);
	size_t usize = m->m_ops->get_user_size(m);

	VALGRIND_DO_MEMPOOL_ALLOC(zone_get_base_address(heap, uptr), uptr, usize);
	VALGRIND_DO_MAKE_MEM_UNDEFINED(uptr, usize);
	VALGRIND_ANNOTATE_NEW_MEMORY(uptr, usize);

	m->m_ops->write_header(m, extra_field, object_flags);

	/*
	 * Set allocated memory with pattern, if debug.heap.alloc_pattern CTL
	 * parameter had been set.
	 */
	if (unlikely(heap->alloc_pattern > PALLOC_CTL_DEBUG_NO_PATTERN)) {
		mo_wal_memset(&heap->p_ops, uptr, heap->alloc_pattern,
			usize, 0);
		VALGRIND_DO_MAKE_MEM_UNDEFINED(uptr, usize);
	}

	int ret;

	if (constructor != NULL) {
		ret = constructor(heap->p_ops.base, uptr, usize, arg);
		if (ret  != 0) {
			/*
			 * If canceled, revert the block back to the free
			 * state in vg machinery.
			 */
			VALGRIND_DO_MEMPOOL_FREE(zone_get_base_address(heap, uptr), uptr);
			return ret;
		}
	}

	/*
	 * To avoid determining the user data pointer twice this method is also
	 * responsible for calculating the offset of the object in the pool that
	 * will be used to set the offset destination pointer provided by the
	 * caller.
	 */
	out->offset = HEAP_PTR_TO_OFF(heap, uptr);
	out->usable_size = usize;

	return 0;
}

/*
 * palloc_reservation_create -- creates a volatile reservation of a
 *	memory block.
 *
 * The first step in the allocation of a new block is reserving it in
 * the transient heap - which is represented by the bucket abstraction.
 *
 * To provide optimal scaling for multi-threaded applications and reduce
 * fragmentation the appropriate bucket is chosen depending on the
 * current thread context and to which allocation class the requested
 * size falls into.
 *
 * Once the bucket is selected, just enough memory is reserved for the
 * requested size. The underlying block allocation algorithm
 * (best-fit, next-fit, ...) varies depending on the bucket container.
 */
static int
palloc_reservation_create(struct palloc_heap *heap, size_t size, palloc_constr constructor,
			  void *arg, uint64_t extra_field, uint16_t object_flags, uint16_t class_id,
			  uint32_t mb_id, struct dav_action_internal *out)
{
	int                  err       = 0;
	struct memory_block *new_block = &out->m;
	struct mbrt         *mb;

	out->type = DAV_ACTION_TYPE_HEAP;

	ASSERT(class_id < UINT8_MAX);
	struct alloc_class *c = class_id == 0 ?
		heap_get_best_class(heap, size) :
		alloc_class_by_id(heap_alloc_classes(heap),
			(uint8_t)class_id);

	if (c == NULL) {
		ERR("no allocation class for size %lu bytes", size);
		errno = EINVAL;
		return -1;
	}

	heap_soemb_active_iter_init(heap);

retry:
	mb = heap_mbrt_get_mb(heap, mb_id);
	if (mb == NULL) {
		errno = EINVAL;
		return -1;
	}

	/*
	 * The caller provided size in bytes, but buckets operate in
	 * 'size indexes' which are multiples of the block size in the
	 * bucket.
	 *
	 * For example, to allocate 500 bytes from a bucket that
	 * provides 256 byte blocks two memory 'units' are required.
	 */
	ssize_t size_idx = alloc_class_calc_size_idx(c, size);

	if (size_idx < 0) {
		ERR("allocation class not suitable for size %lu bytes",
			size);
		errno = EINVAL;
		return -1;
	}
	ASSERT(size_idx <= UINT32_MAX);
	*new_block = MEMORY_BLOCK_NONE;
	new_block->size_idx = (uint32_t)size_idx;

	err = heap_mbrt_update_alloc_class_buckets(heap, mb, c);
	if (err != 0) {
		errno = err;
		return -1;
	}

	struct bucket *b = mbrt_bucket_acquire(mb, c->id);

	err = heap_get_bestfit_block(heap, b, new_block);
	if (err != 0)
		goto out;

	if (alloc_prep_block(heap, new_block, constructor, arg,
		extra_field, object_flags, out) != 0) {
		/*
		 * Constructor returned non-zero value which means
		 * the memory block reservation has to be rolled back.
		 */
		if (new_block->type == MEMORY_BLOCK_HUGE)
			bucket_insert_block(b, new_block);
		err = ECANCELED;
		goto out;
	}

	/*
	 * Each as of yet unfulfilled reservation needs to be tracked in the
	 * runtime state.
	 * The memory block cannot be put back into the global state unless
	 * there are no active reservations.
	 */
	out->mresv = bucket_active_block(b);
	if (out->mresv != NULL)
		util_fetch_and_add64(&out->mresv->nresv, 1);

	out->lock = new_block->m_ops->get_lock(new_block);
	out->new_state = MEMBLOCK_ALLOCATED;

out:
	mbrt_bucket_release(b);

	if (err == 0)
		return 0;

	/*
	 * If there is no memory in evictable zone then do the allocation
	 * from non-evictable zone.
	 */
	if ((mb_id != 0) && (err == ENOMEM)) {
		heap_mbrt_log_alloc_failure(heap, mb_id);
		mb_id = heap_soemb_active_get(heap);
		goto retry;
	}

	errno = err;
	return -1;
}

/*
 * palloc_heap_action_exec -- executes a single heap action (alloc, free)
 */
static void
palloc_heap_action_exec(struct palloc_heap *heap,
	const struct dav_action_internal *act,
	struct operation_context *ctx)
{
	struct zone *zone;
	bool         is_evictable = false;
#ifdef DAV_EXTRA_DEBUG
	if (act->m.m_ops->get_state(&act->m) == act->new_state) {
		D_CRIT("invalid operation or heap corruption\n");
		ASSERT(0);
	}
#endif

	/*
	 * The actual required metadata modifications are chunk-type
	 * dependent, but it always is a modification of a single 8 byte
	 * value - either modification of few bits in a bitmap or
	 * changing a chunk type from free to used or vice versa.
	 */
	act->m.m_ops->prep_hdr(&act->m, act->new_state, ctx);

	/*
	 * Update the memory bucket utilization info.
	 */
	if (heap_mbrt_ismb_evictable(heap, act->m.zone_id))
		is_evictable = true;

	zone = ZID_TO_ZONE(&heap->layout_info, act->m.zone_id);

	if (act->new_state == MEMBLOCK_FREE) {
		zone->header.sp_usage -= act->m.m_ops->get_real_size(&act->m);
		if (!is_evictable && !zone->header.sp_usage)
			heap_incr_empty_nemb_cnt(heap);
	} else {
		if (!is_evictable && !zone->header.sp_usage)
			heap_decr_empty_nemb_cnt(heap);
		zone->header.sp_usage += act->m.m_ops->get_real_size(&act->m);
	}
	operation_add_entry(ctx, &zone->header.sp_usage, zone->header.sp_usage, ULOG_OPERATION_SET);
}

/*
 * palloc_restore_free_chunk_state -- updates the runtime state of a free chunk.
 *
 * This function also takes care of coalescing of huge chunks.
 */
static void
palloc_restore_free_chunk_state(struct palloc_heap *heap,
	struct memory_block *m)
{
	struct mbrt *mb = heap_mbrt_get_mb(heap, m->zone_id);

	if (m->type == MEMORY_BLOCK_HUGE) {
		struct bucket *b = mbrt_bucket_acquire(mb, DEFAULT_ALLOC_CLASS_ID);

		if (heap_free_chunk_reuse(heap, b, m) != 0) {
			if (errno == EEXIST)
				FATAL("duplicate runtime chunk state, possible double free");
			else
				D_CRIT("unable to track runtime chunk state\n");
		}
		mbrt_bucket_release(b);
	}
}

/*
 * palloc_mem_action_noop -- empty handler for unused memory action funcs
 */
static void
palloc_mem_action_noop(struct palloc_heap *heap,
	struct dav_action_internal *act)
{
	/* suppress unused-parameter errors */
	SUPPRESS_UNUSED(heap, act);
}

/*
 * palloc_reservation_clear -- clears the reservation state of the block,
 *	discards the associated memory block if possible
 */
static void
palloc_reservation_clear(struct palloc_heap *heap,
	struct dav_action_internal *act, int publish)
{
	if (act->mresv == NULL)
		return;

	struct memory_block_reserved *mresv = act->mresv;
	struct bucket_locked *locked = mresv->bucket;

	if (!publish) {
		/*
		 * If a memory block used for the action is the currently active
		 * memory block of the bucket it can be returned back to the
		 * bucket. This way it will be available for future allocation
		 * requests, improving performance.
		 */
		struct bucket *b = bucket_acquire(locked);

		bucket_try_insert_attached_block(b, &act->m);
		bucket_release(b);
	}

	if (util_fetch_and_sub64(&mresv->nresv, 1) == 1) {
		VALGRIND_ANNOTATE_HAPPENS_AFTER(&mresv->nresv);
		/*
		 * If the memory block used for the action is not currently used
		 * in any bucket nor action it can be discarded (given back to
		 * the heap).
		 */
		heap_discard_run(heap, &mresv->m);
		D_FREE(mresv);
	} else {
		VALGRIND_ANNOTATE_HAPPENS_BEFORE(&mresv->nresv);
	}
}

/*
 * palloc_heap_action_on_cancel -- restores the state of the heap
 */
static void
palloc_heap_action_on_cancel(struct palloc_heap *heap,
	struct dav_action_internal *act)
{
	void *uptr;

	if (act->new_state == MEMBLOCK_FREE)
		return;

	uptr = act->m.m_ops->get_user_data(&act->m);
	VALGRIND_DO_MEMPOOL_FREE(zone_get_base_address(heap, uptr), uptr);

	act->m.m_ops->invalidate(&act->m);
	palloc_restore_free_chunk_state(heap, &act->m);

	palloc_reservation_clear(heap, act, 0 /* publish */);
}

/*
 * palloc_heap_action_on_process -- performs finalization steps under a lock
 *	on the persistent state
 */
static void
palloc_heap_action_on_process(struct palloc_heap *heap,
	struct dav_action_internal *act)
{
	if (act->new_state == MEMBLOCK_ALLOCATED) {
		STATS_INC(heap->stats, persistent, heap_curr_allocated,
			act->m.m_ops->get_real_size(&act->m));
		if (act->m.type == MEMORY_BLOCK_RUN) {
			STATS_INC(heap->stats, transient, heap_run_allocated,
				act->m.m_ops->get_real_size(&act->m));
		}
		heap_mbrt_incrmb_usage(heap, act->m.zone_id, act->m.m_ops->get_real_size(&act->m));
	} else if (act->new_state == MEMBLOCK_FREE) {
		if (On_memcheck) {
			void *ptr = act->m.m_ops->get_user_data(&act->m);

			VALGRIND_DO_MEMPOOL_FREE(zone_get_base_address(heap, ptr), ptr);
		}

		STATS_SUB(heap->stats, persistent, heap_curr_allocated,
			act->m.m_ops->get_real_size(&act->m));
		if (act->m.type == MEMORY_BLOCK_RUN) {
			STATS_SUB(heap->stats, transient, heap_run_allocated,
				act->m.m_ops->get_real_size(&act->m));
		}
		heap_memblock_on_free(heap, &act->m);
		heap_mbrt_incrmb_usage(heap, act->m.zone_id,
				       -(act->m.m_ops->get_real_size(&act->m)));
	}
}

/*
 * palloc_heap_action_on_unlock -- performs finalization steps that need to be
 *	performed without a lock on persistent state
 */
static void
palloc_heap_action_on_unlock(struct palloc_heap *heap,
	struct dav_action_internal *act)
{
	if (act->new_state == MEMBLOCK_ALLOCATED)
		palloc_reservation_clear(heap, act, 1 /* publish */);
	else if (act->new_state == MEMBLOCK_FREE)
		palloc_restore_free_chunk_state(heap, &act->m);
}

/*
 * palloc_mem_action_exec -- executes a single memory action (set, and, or)
 */
static void
palloc_mem_action_exec(struct palloc_heap *heap,
	const struct dav_action_internal *act,
	struct operation_context *ctx)
{
	/* suppress unused-parameter errors */
	SUPPRESS_UNUSED(heap);

	operation_add_entry(ctx, act->ptr, act->value, ULOG_OPERATION_SET);
}

static const struct {
	/*
	 * Translate action into some number of operation_entry'ies.
	 */
	void (*exec)(struct palloc_heap *heap,
		const struct dav_action_internal *act,
		struct operation_context *ctx);

	/*
	 * Cancel any runtime state changes. Can be called only when action has
	 * not been translated to persistent operation yet.
	 */
	void (*on_cancel)(struct palloc_heap *heap,
		struct dav_action_internal *act);

	/*
	 * Final steps after persistent state has been modified. Performed
	 * under action-specific lock.
	 */
	void (*on_process)(struct palloc_heap *heap,
		struct dav_action_internal *act);

	/*
	 * Final steps after persistent state has been modified. Performed
	 * after action-specific lock has been dropped.
	 */
	void (*on_unlock)(struct palloc_heap *heap,
		struct dav_action_internal *act);
} action_funcs[DAV_MAX_ACTION_TYPE] = {
	[DAV_ACTION_TYPE_HEAP] = {
		.exec = palloc_heap_action_exec,
		.on_cancel = palloc_heap_action_on_cancel,
		.on_process = palloc_heap_action_on_process,
		.on_unlock = palloc_heap_action_on_unlock,
	},
	[DAV_ACTION_TYPE_MEM] = {
		.exec = palloc_mem_action_exec,
		.on_cancel = palloc_mem_action_noop,
		.on_process = palloc_mem_action_noop,
		.on_unlock = palloc_mem_action_noop,
	}
};

/*
 * palloc_action_compare -- compares two actions based on lock address
 */
static int
palloc_action_compare(const void *lhs, const void *rhs)
{
	const struct dav_action_internal *mlhs = lhs;
	const struct dav_action_internal *mrhs = rhs;
	uintptr_t vlhs = (uintptr_t)(mlhs->lock);
	uintptr_t vrhs = (uintptr_t)(mrhs->lock);

	if (vlhs < vrhs)
		return -1;
	if (vlhs > vrhs)
		return 1;

	return 0;
}

/*
 * palloc_exec_actions -- perform the provided free/alloc operations
 */
static void
palloc_exec_actions(struct palloc_heap *heap,
	struct operation_context *ctx,
	struct dav_action_internal *actv,
	size_t actvcnt)
{
	/*
	 * The operations array is sorted so that proper lock ordering is
	 * ensured.
	 */
	if (actv)
		qsort(actv, actvcnt, sizeof(struct dav_action_internal),
			palloc_action_compare);
	else
		ASSERTeq(actvcnt, 0);

	struct dav_action_internal *act;

	for (size_t i = 0; i < actvcnt; ++i) {
		act = &actv[i];

		/*
		 * This lock must be held for the duration between the creation
		 * of the allocation metadata updates in the operation context
		 * and the operation processing. This is because a different
		 * thread might operate on the same 8-byte value of the run
		 * bitmap and override allocation performed by this thread.
		 */
		if (i == 0 || act->lock != actv[i - 1].lock) {
			if (act->lock)
				util_mutex_lock(act->lock);
		}

		/* translate action to some number of operation_entry'ies */
		action_funcs[act->type].exec(heap, act, ctx);
	}

	/* wait for all allocated object headers to be persistent */
	mo_wal_drain(&heap->p_ops);

	/* perform all persistent memory operations */
	operation_process(ctx);

	for (size_t i = 0; i < actvcnt; ++i) {
		act = &actv[i];

		action_funcs[act->type].on_process(heap, act);

		if (i == actvcnt - 1 || act->lock != actv[i + 1].lock) {
			if (act->lock)
				util_mutex_unlock(act->lock);
		}
	}

	for (size_t i = 0; i < actvcnt; ++i) {
		act = &actv[i];

		action_funcs[act->type].on_unlock(heap, act);
	}

	operation_finish(ctx, 0);
}

/*
 * palloc_reserve -- creates a single reservation
 */
int
palloc_reserve(struct palloc_heap *heap, size_t size, palloc_constr constructor, void *arg,
	       uint64_t extra_field, uint16_t object_flags, uint16_t class_id, uint32_t mb_id,
	       struct dav_action *act)
{
	COMPILE_ERROR_ON(sizeof(struct dav_action) !=
		sizeof(struct dav_action_internal));

	return palloc_reservation_create(heap, size, constructor, arg, extra_field, object_flags,
					 class_id, mb_id, (struct dav_action_internal *)act);
}

/*
 * palloc_action_isalloc - action is a heap reservation
 *			   created by palloc_reserve().
 */
int
palloc_action_isalloc(struct dav_action *act)
{
	struct dav_action_internal *actp = (struct dav_action_internal *)act;

	return ((actp->type == DAV_ACTION_TYPE_HEAP) &&
		(actp->new_state == MEMBLOCK_ALLOCATED));
}

uint64_t
palloc_get_realoffset(struct palloc_heap *heap, uint64_t off)
{
	struct memory_block m = memblock_from_offset(heap, off);

	return HEAP_PTR_TO_OFF(m.heap, m.m_ops->get_real_data(&m));
}

/*
 * palloc_get_prange -- get the start offset and size of allocated memory that
 *			needs to be persisted.
 *
 * persist_udata - if true, persist the user data.
 */
void
palloc_get_prange(struct dav_action *act, uint64_t *const offp, uint64_t *const sizep,
		  int persist_udata)
{
	struct dav_action_internal *act_in = (struct dav_action_internal *)act;

	D_ASSERT(act_in->type == DAV_ACTION_TYPE_HEAP);
	/* we need to persist the header if present */
	*offp = HEAP_PTR_TO_OFF(act_in->m.heap, act_in->m.m_ops->get_real_data(&act_in->m));
	*sizep = header_type_to_size[act_in->m.header_type];

	D_ASSERT(act_in->offset == *offp + header_type_to_size[act_in->m.header_type]);
	/* persist the user data */
	if (persist_udata)
		*sizep += act_in->usable_size;
}

/*
 * palloc_defer_free -- creates an internal deferred free action
 */
static void
palloc_defer_free_create(struct palloc_heap *heap, uint64_t off,
			 struct dav_action_internal *out)
{
	COMPILE_ERROR_ON(sizeof(struct dav_action) !=
		sizeof(struct dav_action_internal));

	out->type = DAV_ACTION_TYPE_HEAP;
	out->offset = off;
	out->m = memblock_from_offset(heap, off);

	/*
	 * For the duration of free we may need to protect surrounding
	 * metadata from being modified.
	 */
	out->lock = out->m.m_ops->get_lock(&out->m);
	out->mresv = NULL;
	out->new_state = MEMBLOCK_FREE;
}

/*
 * palloc_defer_free -- creates a deferred free action
 */
void
palloc_defer_free(struct palloc_heap *heap, uint64_t off, struct dav_action *act)
{
	COMPILE_ERROR_ON(sizeof(struct dav_action) !=
		sizeof(struct dav_action_internal));

	palloc_defer_free_create(heap, off, (struct dav_action_internal *)act);
}

/*
 * palloc_cancel -- cancels all reservations in the array
 */
void
palloc_cancel(struct palloc_heap *heap, struct dav_action *actv, size_t actvcnt)
{
	struct dav_action_internal *act;

	for (size_t i = 0; i < actvcnt; ++i) {
		act = (struct dav_action_internal *)&actv[i];
		action_funcs[act->type].on_cancel(heap, act);
	}
}

/*
 * palloc_publish -- publishes all reservations in the array
 */
void
palloc_publish(struct palloc_heap *heap, struct dav_action *actv, size_t actvcnt,
	       struct operation_context *ctx)
{
	palloc_exec_actions(heap, ctx,
		(struct dav_action_internal *)actv, actvcnt);
}

/*
 * palloc_operation -- persistent memory operation. Takes a NULL pointer
 *	or an existing memory block and modifies it to occupy, at least, 'size'
 *	number of bytes.
 *
 * The malloc, free and realloc routines are implemented in the context of this
 * common operation which encompasses all of the functionality usually done
 * separately in those methods.
 *
 * The first thing that needs to be done is determining which memory blocks
 * will be affected by the operation - this varies depending on the whether the
 * operation will need to modify or free an existing block and/or allocate
 * a new one.
 *
 * Simplified allocation process flow is as follows:
 *	- reserve a new block in the transient heap
 *	- prepare the new block
 *	- create redo log of required modifications
 *		- chunk metadata
 *		- offset of the new object
 *	- commit and process the redo log
 *
 * And similarly, the deallocation process:
 *	- create redo log of required modifications
 *		- reverse the chunk metadata back to the 'free' state
 *		- set the destination of the object offset to zero
 *	- commit and process the redo log
 * There's an important distinction in the deallocation process - it does not
 * return the memory block to the transient container. That is done once no more
 * memory is available.
 *
 * Reallocation is a combination of the above, with one additional step
 * of copying the old content.
 */
int
palloc_operation(struct palloc_heap *heap, uint64_t off, uint64_t *dest_off, size_t size,
		 palloc_constr constructor, void *arg, uint64_t extra_field, uint16_t object_flags,
		 uint16_t class_id, uint32_t mb_id, struct operation_context *ctx)
{
	size_t user_size = 0;

	size_t nops = 0;
	uint64_t aoff;
	uint64_t asize;
	struct dav_action_internal ops[2];
	struct dav_action_internal *alloc = NULL;
	struct dav_action_internal *dealloc = NULL;

	/*
	 * The offset of an existing block can be nonzero which means this
	 * operation is either free or a realloc - either way the offset of the
	 * object needs to be translated into memory block, which is a structure
	 * that all of the heap methods expect.
	 */
	if (off != 0) {
		dealloc = &ops[nops++];
		palloc_defer_free_create(heap, off, dealloc);
		user_size = dealloc->m.m_ops->get_user_size(&dealloc->m);
		if (user_size == size) {
			operation_cancel(ctx);
			return 0;
		}
	}

	/* alloc or realloc */
	if (size != 0) {
		alloc = &ops[nops++];
		if (palloc_reservation_create(heap, size, constructor, arg, extra_field,
					      object_flags, class_id, mb_id, alloc) != 0) {
			operation_cancel(ctx);
			return -1;
		}

		palloc_get_prange((struct dav_action *)alloc, &aoff, &asize, 0);
		if (asize) /* != CHUNK_FLAG_HEADER_NONE */
			dav_wal_tx_snap(heap->p_ops.base, HEAP_OFF_TO_PTR(heap, aoff),
					asize, HEAP_OFF_TO_PTR(heap, aoff), 0);
	}

	/* realloc */
	if (alloc != NULL && dealloc != NULL) {
		/* copy data to newly allocated memory */
		size_t old_size = user_size;
		size_t to_cpy = old_size > size ? size : old_size;

		VALGRIND_ADD_TO_TX(
			HEAP_OFF_TO_PTR(heap, alloc->offset),
			to_cpy);
		mo_wal_memcpy(&heap->p_ops,
			HEAP_OFF_TO_PTR(heap, alloc->offset),
			HEAP_OFF_TO_PTR(heap, off),
			to_cpy,
			0);
		VALGRIND_REMOVE_FROM_TX(
			HEAP_OFF_TO_PTR(heap, alloc->offset),
			to_cpy);
	}

	/*
	 * If the caller provided a destination value to update, it needs to be
	 * modified atomically alongside the heap metadata, and so the operation
	 * context must be used.
	 */
	if (dest_off) {
		operation_add_entry(ctx, dest_off,
			alloc ? alloc->offset : 0, ULOG_OPERATION_SET);
	}

	/* and now actually perform the requested operation! */
	palloc_exec_actions(heap, ctx, ops, nops);

	return 0;
}

/*
 * palloc_usable_size -- returns the number of bytes in the memory block
 */
size_t
palloc_usable_size(struct palloc_heap *heap, uint64_t off)
{
	struct memory_block m = memblock_from_offset(heap, off);

	return m.m_ops->get_user_size(&m);
}

/*
 * palloc_extra -- returns allocation extra field
 */
uint64_t
palloc_extra(struct palloc_heap *heap, uint64_t off)
{
	struct memory_block m = memblock_from_offset(heap, off);

	return m.m_ops->get_extra(&m);
}

/*
 * palloc_flags -- returns allocation flags
 */
uint16_t
palloc_flags(struct palloc_heap *heap, uint64_t off)
{
	struct memory_block m = memblock_from_offset(heap, off);

	return m.m_ops->get_flags(&m);
}

/*
 * pmalloc_search_cb -- (internal) foreach callback.
 */
static int
pmalloc_search_cb(const struct memory_block *m, void *arg)
{
	struct memory_block *out = arg;

	if (MEMORY_BLOCK_EQUALS(*m, *out))
		return 0; /* skip the same object */

	*out = *m;

	return 1;
}

/*
 * palloc_first -- returns the first object from the heap.
 */
uint64_t
palloc_first(struct palloc_heap *heap)
{
	struct memory_block search = MEMORY_BLOCK_NONE;

	heap_foreach_object(heap, pmalloc_search_cb,
		&search, MEMORY_BLOCK_NONE);

	if (MEMORY_BLOCK_IS_NONE(search))
		return 0;

	void *uptr = search.m_ops->get_user_data(&search);

	return HEAP_PTR_TO_OFF(heap, uptr);
}

/*
 * palloc_next -- returns the next object relative to 'off'.
 */
uint64_t
palloc_next(struct palloc_heap *heap, uint64_t off)
{
	struct memory_block m = memblock_from_offset(heap, off);
	struct memory_block search = m;

	heap_foreach_object(heap, pmalloc_search_cb, &search, m);

	if (MEMORY_BLOCK_IS_NONE(search) ||
		MEMORY_BLOCK_EQUALS(search, m))
		return 0;

	void *uptr = search.m_ops->get_user_data(&search);

	return HEAP_PTR_TO_OFF(heap, uptr);
}

#if VG_MEMCHECK_ENABLED
/*
 * palloc_vg_register_alloc -- (internal) registers allocation header
 * in Valgrind
 */
static int
palloc_vg_register_alloc(const struct memory_block *m, void *arg)
{
	struct palloc_heap *heap = arg;

	m->m_ops->reinit_header(m);

	void *uptr = m->m_ops->get_user_data(m);
	size_t usize = m->m_ops->get_user_size(m);

	VALGRIND_DO_MEMPOOL_ALLOC(zone_get_base_address(heap, uptr), uptr, usize);
	VALGRIND_DO_MAKE_MEM_DEFINED(uptr, usize);

	return 0;
}

/*
 * palloc_heap_vg_open -- notifies Valgrind about heap layout
 */
void
palloc_heap_vg_open(struct palloc_heap *heap, int objects)
{
	heap_vg_open(heap, palloc_vg_register_alloc, heap, objects);
}

void
palloc_heap_vg_zone_open(struct palloc_heap *heap, uint32_t zid, int objects)
{
	heap_vg_zone_open(heap, zid, palloc_vg_register_alloc, heap, objects);
}
#endif
