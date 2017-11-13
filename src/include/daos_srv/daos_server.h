/**
 * (C) Copyright 2016 Intel Corporation.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * GOVERNMENT LICENSE RIGHTS-OPEN SOURCE SOFTWARE
 * The Government's rights to use, modify, reproduce, release, perform, display,
 * or disclose this software are subject to the terms of the Apache License as
 * provided in Contract No. B609815.
 * Any reproduction of computer software, computer software documentation, or
 * portions thereof marked with this legend must also reproduce the markings.
 */
/**
 * DAOS server-side infrastructure
 * Provides a modular interface to load server-side code on demand
 */

#ifndef __DSS_API_H__
#define __DSS_API_H__

#include <daos/common.h>
#include <daos/rpc.h>
#include <daos_srv/iv.h>
#include <daos_event.h>
#include <daos_task.h>

#include <pthread.h>
#include <hwloc.h>
#include <abt.h>
#include <cart/iv.h>

/** Number of execution streams started or cores used */
extern unsigned int	dss_nxstreams;

/** Server node topoloby */
extern hwloc_topology_t	dss_topo;

/** Storage path (hack) */
extern const char      *storage_path;

/**
 * Stackable Module API
 * Provides a modular interface to load and register server-side code on
 * demand. A module is composed of:
 * - a set of request handlers which are registered when the module is loaded.
 * - a server-side API (see header files suffixed by "_srv") used for
 *   inter-module direct calls.
 *
 * For now, all loaded modules are assumed to be trustful, but sandboxes can be
 * implemented in the future.
 */
/*
 * Thead-local storage
 */
struct dss_thread_local_storage {
	uint32_t	dtls_tag;
	void		**dtls_values;
};

enum dss_module_tag {
	/* Server tag */
	DAOS_SERVER_TAG	= 1 << 0,
};

/* The module key descriptor for each server thread */
struct dss_module_key {
	/* Indicate where the keys should be instantiated */
	enum dss_module_tag dmk_tags;

	/* The position inside the dss_module_keys */
	int dmk_index;
	/* init keys for context */
	void  *(*dmk_init)(const struct dss_thread_local_storage *dtls,
			   struct dss_module_key *key);

	/* fini keys for context */
	void  (*dmk_fini)(const struct dss_thread_local_storage *dtls,
			  struct dss_module_key *key, void *data);
};

extern pthread_key_t dss_tls_key;
extern struct dss_module_key *dss_module_keys[];
#define DAOS_MODULE_KEYS_NR 10

static inline struct dss_thread_local_storage *
dss_tls_get()
{
	return (struct dss_thread_local_storage *)
		pthread_getspecific(dss_tls_key);
}

/**
 * Get value from context by the key
 *
 * Get value inside dtls by key. So each module will use this API to
 * retrieve their own value in the thread context.
 *
 * \param[in] dtls	the thread context.
 * \param[in] key	key used to retrieve the dtls_value.
 *
 * \retval		the dtls_value retrieved by key.
 */
static inline void *
dss_module_key_get(struct dss_thread_local_storage *dtls,
		   struct dss_module_key *key)
{
	D__ASSERT(key->dmk_index >= 0);
	D__ASSERT(key->dmk_index < DAOS_MODULE_KEYS_NR);
	D__ASSERT(dss_module_keys[key->dmk_index] == key);

	return dtls->dtls_values[key->dmk_index];
}

void dss_register_key(struct dss_module_key *key);
void dss_unregister_key(struct dss_module_key *key);

struct dss_module_info {
	crt_context_t		dmi_ctx;
	struct dss_xstream	*dmi_xstream;
	int			dmi_tid;
	tse_sched_t		dmi_sched;
};

extern struct dss_module_key	daos_srv_modkey;

static inline struct dss_module_info *
dss_get_module_info()
{
	struct dss_module_info *dmi;
	struct dss_thread_local_storage *dtc;

	dtc = dss_tls_get();
	dmi = (struct dss_module_info *)
	      dss_module_key_get(dtc, &daos_srv_modkey);
	return dmi;
}

/**
 * Module facility feature bits
 * DSS_FAC_LOAD_CLI - the module requires loading client stack.
 */
#define DSS_FAC_LOAD_CLI (0x1ULL)

/**
 * Each module should provide a dss_module structure which defines the module
 * interface. The name of the allocated structure must be the library name
 * (without the ".so" extension) suffixed by "module". This symbol will be
 * looked up automatically when the module library is loaded and failed if not
 * found.
 *
 * For instance, the dmg module reports a "sm_name" of "daos_mgmt_srv", the
 * actual library filename is libdaos_mgmt_srv.so and it defines a dss_module
 * structure called daos_mgmt_srv_module.
 */
struct dss_module {
	/* Name of the module */
	const char		 *sm_name;
	/* Module id see enum dss_module_id */
	int			  sm_mod_id;
	/* Module version */
	int			  sm_ver;
	/* Module facility bitmask, can be feature bits like DSS_FAC_LOAD_CLI */
	uint64_t		  sm_facs;
	/* key of local thread storage */
	struct dss_module_key	*sm_key;
	/* Setup function, invoked just after successful load */
	int			(*sm_init)(void);
	/* Teardown function, invoked just before module unload */
	int			(*sm_fini)(void);
	/* ES setup function, invoked for each ES */
	int			(*sm_xstream_init)(void);
	/* ES teardown function, invoked for each ES */
	int			(*sm_xstream_fini)(void);
	/* Array of RPC definition for request sent by client nodes, last entry
	 * of the array must be empty */
	struct daos_rpc		 *sm_cl_rpcs;
	/* Array of RPC definition for request sent by other servers, last entry
	 * of the array must be empty */
	struct daos_rpc		 *sm_srv_rpcs;

	/* RPC handler of these RPC, last entry of the array must be empty */
	struct daos_rpc_handler	 *sm_handlers;
};

int dss_ult_create(void (*func)(void *), void *arg,
		   int stream_id, ABT_thread *ult);
int dss_ult_create_all(void (*func)(void *), void *arg);
int dss_ult_create_execute(int (*func)(void *), void *arg,
			   int stream_id);

/* Pack return codes with additional argument to reduce */
struct dss_stream_arg_type {
	/** return value */
	int		st_rc;
	/** collective arguments for streams */
	void		*st_coll_args;
	/** optional reduce args for aggregation */
	void		*st_arg;
};

struct dss_coll_stream_args {
	struct dss_stream_arg_type *csa_streams;
};

struct dss_coll_ops {
	/**
	 * Function to be invoked by dss_collective
	 *
	 * \param f_args		[IN]	Arguments for function
	 */
	int				(*co_func)(void *f_args);

	/**
	 * Callback for reducing after dss_collective (optional)
	 *
	 * \param a_args		[IN/OUT]
	 *					Aggregator arguments for
	 *					reducing results
	 * \param s_args		[IN]	Reduce arguments for this
	 *					current stream
	 */
	void				(*co_reduce)(void *a_args,
						     void *s_args);

	/**
	 * Alloc function for allocating reduce arguments (optional)
	 *
	 * \param args			[IN/OUT] coll_args for this streams
	 * \param aggregator_args	[IN]	 aggregator args for
	 *					 initializatuin
	 */
	void				(*co_reduce_arg_alloc)
					(struct dss_stream_arg_type *args,
					 void *a_args);
	/**
	 * Free the allocated reduce arguments
	 * (Mandatory if co_rarg_alloc was provided)
	 *
	 * \param args			[IN]	coll_args for this stream
	 */
	void				(*co_reduce_arg_free)
					(struct dss_stream_arg_type *args);
};

struct dss_coll_args {
	/** Arguments for dss_collective func (Mandatory) */
	void				*ca_func_args;
	void				*ca_aggregator;
	/** Stream arguments for all streams */
	struct dss_coll_stream_args	ca_stream_args;
};

/* Generic dss_collective with custom aggregator */
int
dss_task_collective_reduce(struct dss_coll_ops *ops,
			   struct dss_coll_args *coll_args);
int
dss_thread_collective_reduce(struct dss_coll_ops *ops,
			     struct dss_coll_args *coll_args);

int dss_task_collective(int (*func)(void *), void *arg);
int dss_thread_collective(int (*func)(void *), void *arg);

int dss_sync_task(daos_opc_t opc, void *arg, unsigned int arg_size);
unsigned int dss_get_threads_number(void);

/* Convert Argobots errno to DAOS ones. */
static inline int
dss_abterr2der(int abt_errno)
{
	switch (abt_errno) {
	case ABT_SUCCESS:	return 0;
	case ABT_ERR_MEM:	return -DER_NOMEM;
	default:		return -DER_INVAL;
	}
}

int dss_rpc_send(crt_rpc_t *rpc);
int dss_group_create(crt_group_id_t id, d_rank_list_t *ranks,
		     crt_group_t **group);
int dss_group_destroy(crt_group_t *group);

void dss_sleep(int ms);

enum iv_key {
	IV_POOL_MAP = 1,
	IV_REBUILD,
};

int ds_iv_fetch(struct ds_iv_ns *ns, unsigned int key_id,
		d_sg_list_t *value);
int ds_iv_update(struct ds_iv_ns *ns, unsigned int key_id,
		 d_sg_list_t *value, unsigned int shortcut,
		 unsigned int sync_mode);
int ds_iv_invalidate(struct ds_iv_ns *ns, unsigned int key_id,
		     unsigned int shortcut, unsigned int sync_mode);

int ds_iv_ns_create(crt_context_t ctx, unsigned int *class_id,
		    daos_iov_t *g_ivns, struct ds_iv_ns **p_iv_ns);

int ds_iv_ns_attach(crt_context_t ctx, unsigned int class_id,
		    unsigned int master_rank, daos_iov_t *iv_ctxt,
		    struct ds_iv_ns **p_iv_ns);

void ds_iv_ns_destroy(void *ns);

unsigned int ds_iv_ns_id_get(void *ns);
#endif /* __DSS_API_H__ */
