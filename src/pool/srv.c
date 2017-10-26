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
 * ds_pool: Pool Server
 *
 * This is part of daos_server. It exports the pool RPC handlers and implements
 * Pool Server API.
 */
#define DDSUBSYS	DDFAC(pool)

#include <daos_srv/pool.h>

#include <daos/rpc.h>
#include <daos_srv/daos_mgmt_srv.h>
#include <daos_srv/daos_server.h>
#include "rpc.h"
#include "srv_internal.h"

static int
init(void)
{
	int rc;

	rc = ds_pool_svc_hash_init();
	if (rc != 0)
		D__GOTO(err, rc);

	rc = ds_pool_cache_init();
	if (rc != 0)
		D__GOTO(err_pool_svc, rc);

	rc = ds_pool_hdl_hash_init();
	if (rc != 0)
		D__GOTO(err_pool_cache, rc);

	rc = ds_pool_iv_init();
	if (rc)
		D_GOTO(err_hdl_hash, rc);
	return 0;
err_hdl_hash:
	ds_pool_hdl_hash_fini();
err_pool_cache:
	ds_pool_cache_fini();
err_pool_svc:
	ds_pool_svc_hash_fini();
err:
	return rc;
}

static int
fini(void)
{
	ds_pool_hdl_hash_fini();
	ds_pool_cache_fini();
	ds_pool_svc_hash_fini();
	ds_pool_iv_fini();
	return 0;
}

/*
 * Try to start a pool's pool service if its RDB exists. Continue the iteration
 * upon errors as other pools may still be able to work.
 */
static int
start_pool_svc(const uuid_t uuid, void *arg)
{
	char   *path;
	int	rc;

	/*
	 * Check if an RDB file exists and we can access it, to avoid
	 * unnecessary error messages from the ds_pool_svc_start() call.
	 */
	path = ds_pool_rdb_path(uuid, uuid);
	if (path == NULL) {
		D_ERROR(DF_UUID": failed allocate rdb path\n", DP_UUID(uuid));
		return 0;
	}
	rc = access(path, R_OK | W_OK);
	free(path);
	if (rc != 0) {
		D_DEBUG(DB_MD, DF_UUID": cannot find or access rdb: %d\n",
			DP_UUID(uuid), errno);
		return 0;
	}

	rc = ds_pool_svc_start(uuid);
	if (rc != 0) {
		D_ERROR("not starting pool service "DF_UUID": %d\n",
			DP_UUID(uuid), rc);
		return 0;
	}

	D_DEBUG(DB_MD, "started pool service "DF_UUID"\n", DP_UUID(uuid));
	return 0;
}

static int
xstream_init(void)
{
	struct dss_module_info *info = dss_get_module_info();
	int			rc;

	if (info->dmi_tid != 0)
		return 0;

	/* Scan the storage and start all pool services. */
	rc = ds_mgmt_tgt_pool_iterate(start_pool_svc, NULL /* arg */);
	if (rc != 0)
		return rc;

	return 0;
}

static int
xstream_fini(void)
{
	struct dss_module_info *info = dss_get_module_info();

	if (info->dmi_tid != 0)
		return 0;

	/* TODO: Stop all pool services. */

	return 0;
}

/* Note: the rpc input/output parameters is defined in daos_rpc */
static struct daos_rpc_handler pool_handlers[] = {
	{
		.dr_opc		= POOL_CREATE,
		.dr_hdlr	= ds_pool_create_handler
	}, {
		.dr_opc		= POOL_CONNECT,
		.dr_hdlr	= ds_pool_connect_handler
	}, {
		.dr_opc		= POOL_DISCONNECT,
		.dr_hdlr	= ds_pool_disconnect_handler
	}, {
		.dr_opc		= POOL_QUERY,
		.dr_hdlr	= ds_pool_query_handler
	}, {
		.dr_opc		= POOL_EXCLUDE,
		.dr_hdlr	= ds_pool_update_handler
	}, {
		.dr_opc		= POOL_EVICT,
		.dr_hdlr	= ds_pool_evict_handler
	}, {
		.dr_opc		= POOL_ADD,
		.dr_hdlr	= ds_pool_update_handler
	},{
		.dr_opc		= POOL_EXCLUDE_OUT,
		.dr_hdlr	= ds_pool_update_handler
	},{
		.dr_opc		= POOL_SVC_STOP,
		.dr_hdlr	= ds_pool_svc_stop_handler
	}, {
		.dr_opc		= POOL_TGT_CONNECT,
		.dr_hdlr	= ds_pool_tgt_connect_handler,
		.dr_corpc_ops	= {
			.co_aggregate	= ds_pool_tgt_connect_aggregator
		}
	}, {
		.dr_opc		= POOL_TGT_DISCONNECT,
		.dr_hdlr	= ds_pool_tgt_disconnect_handler,
		.dr_corpc_ops	= {
			.co_aggregate	= ds_pool_tgt_disconnect_aggregator
		}
	}, {
		.dr_opc		= POOL_TGT_UPDATE_MAP,
		.dr_hdlr	= ds_pool_tgt_update_map_handler,
		.dr_corpc_ops	= {
			.co_aggregate	= ds_pool_tgt_update_map_aggregator
		}
	}, {
		.dr_opc		= 0
	}
};

static void *
pool_tls_init(const struct dss_thread_local_storage *dtls,
	      struct dss_module_key *key)
{
	struct pool_tls *tls;

	D__ALLOC_PTR(tls);
	if (tls == NULL)
		return NULL;

	DAOS_INIT_LIST_HEAD(&tls->dt_pool_list);
	return tls;
}

static void
pool_tls_fini(const struct dss_thread_local_storage *dtls,
	      struct dss_module_key *key, void *data)
{
	struct pool_tls *tls = data;

	ds_pool_child_purge(tls);
	D__ASSERT(daos_list_empty(&tls->dt_pool_list));
	D__FREE_PTR(tls);
}

struct dss_module_key pool_module_key = {
	.dmk_tags = DAOS_SERVER_TAG,
	.dmk_index = -1,
	.dmk_init = pool_tls_init,
	.dmk_fini = pool_tls_fini,
};

struct dss_module pool_module =  {
	.sm_name		= "pool",
	.sm_mod_id		= DAOS_POOL_MODULE,
	.sm_ver			= 1,
	.sm_init		= init,
	.sm_fini		= fini,
	.sm_xstream_init	= xstream_init,
	.sm_xstream_fini	= xstream_fini,
	.sm_cl_rpcs		= pool_rpcs,
	.sm_srv_rpcs		= pool_srv_rpcs,
	.sm_handlers		= pool_handlers,
	.sm_key			= &pool_module_key,
};
