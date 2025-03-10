/* SPDX-License-Identifier: BSD-3-Clause */
/* Copyright 2018-2023, Intel Corporation */

/*
 * container_ravl.c -- implementation of ravl-based block container
 */

#include "container.h"
#include "ravl.h"
#include "out.h"
#include "sys_util.h"

struct block_container_ravl {
	struct block_container super;
	struct memory_block    m;
	struct ravl           *tree;
};

/*
 * container_compare_memblocks -- (internal) compares two memory blocks
 */
static int
container_compare_memblocks(const void *lhs, const void *rhs)
{
	const struct memory_block *l = lhs;
	const struct memory_block *r = rhs;

	int64_t diff = (int64_t)l->size_idx - (int64_t)r->size_idx;

	if (diff != 0)
		return diff > 0 ? 1 : -1;

	diff = (int64_t)l->zone_id - (int64_t)r->zone_id;
	if (diff != 0)
		return diff > 0 ? 1 : -1;

	diff = (int64_t)l->chunk_id - (int64_t)r->chunk_id;
	if (diff != 0)
		return diff > 0 ? 1 : -1;

	diff = (int64_t)l->block_off - (int64_t)r->block_off;
	if (diff != 0)
		return diff > 0 ? 1 : -1;

	return 0;
}

/*
 * container_ravl_insert_block -- (internal) inserts a new memory block
 *	into the container
 */
static int
container_ravl_insert_block(struct block_container *bc,
	const struct memory_block *m)
{
	struct block_container_ravl *c =
		(struct block_container_ravl *)bc;

	ASSERT(m->chunk_id < MAX_CHUNK);
	ASSERT(m->zone_id < UINT32_MAX);

	c->m = *m;

	return ravl_emplace_copy(c->tree, m);
}

/*
 * container_ravl_get_rm_block_bestfit -- (internal) removes and returns the
 *	best-fit memory block for size
 */
static int
container_ravl_get_rm_block_bestfit(struct block_container *bc,
	struct memory_block *m)
{
	struct block_container_ravl *c =
		(struct block_container_ravl *)bc;

	struct ravl_node *n = ravl_find(c->tree, m,
		RAVL_PREDICATE_GREATER_EQUAL);

	if (n == NULL)
		return ENOMEM;

	struct memory_block *e = ravl_data(n);
	*m                     = c->m;
	m->zone_id             = e->zone_id;
	m->chunk_id            = e->chunk_id;
	m->size_idx            = e->size_idx;
	m->block_off           = e->block_off;
	/* Rest of the fields in e should not be accessed. */

	ravl_remove(c->tree, n);

	return 0;
}

/*
 * container_ravl_get_rm_block_exact --
 *	(internal) removes exact match memory block
 */
static int
container_ravl_get_rm_block_exact(struct block_container *bc,
	const struct memory_block *m)
{
	struct block_container_ravl *c =
		(struct block_container_ravl *)bc;

	struct ravl_node *n = ravl_find(c->tree, m, RAVL_PREDICATE_EQUAL);

	if (n == NULL)
		return ENOMEM;

	ravl_remove(c->tree, n);

	return 0;
}

/*
 * container_ravl_is_empty -- (internal) checks whether the container is empty
 */
static int
container_ravl_is_empty(struct block_container *bc)
{
	struct block_container_ravl *c =
		(struct block_container_ravl *)bc;

	return ravl_empty(c->tree);
}

/*
 * container_ravl_rm_all -- (internal) removes all elements from the tree
 */
static void
container_ravl_rm_all(struct block_container *bc)
{
	struct block_container_ravl *c =
		(struct block_container_ravl *)bc;

	ravl_clear(c->tree);
}

/*
 * container_ravl_delete -- (internal) deletes the container
 */
static void
container_ravl_destroy(struct block_container *bc)
{
	struct block_container_ravl *c =
		(struct block_container_ravl *)bc;

	ravl_delete(c->tree);

	D_FREE(bc);
}

/*
 * Tree-based block container used to provide best-fit functionality to the
 * bucket. The time complexity for this particular container is O(k) where k is
 * the length of the key.
 *
 * The get methods also guarantee that the block with lowest possible address
 * that best matches the requirements is provided.
 */
static const struct block_container_ops container_ravl_ops = {
	.insert = container_ravl_insert_block,
	.get_rm_exact = container_ravl_get_rm_block_exact,
	.get_rm_bestfit = container_ravl_get_rm_block_bestfit,
	.is_empty = container_ravl_is_empty,
	.rm_all = container_ravl_rm_all,
	.destroy = container_ravl_destroy,
};

/*
 * container_new_ravl -- allocates and initializes a ravl container
 */
struct block_container *
container_new_ravl(struct palloc_heap *heap)
{
	struct block_container_ravl *bc;

	D_ALLOC_PTR_NZ(bc);
	if (bc == NULL)
		goto error_container_malloc;

	bc->super.heap = heap;
	bc->super.c_ops = &container_ravl_ops;
	bc->tree =
	    ravl_new_sized(container_compare_memblocks, offsetof(struct memory_block, m_ops));
	if (bc->tree == NULL)
		goto error_ravl_new;

	return (struct block_container *)&bc->super;

error_ravl_new:
	D_FREE(bc);

error_container_malloc:
	return NULL;
}
