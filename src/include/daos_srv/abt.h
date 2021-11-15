/*
 * (C) Copyright 2021 Intel Corporation.
 *
 * SPDX-License-Identifier: BSD-2-Clause-Patent
 */
/**
 * \file
 *
 * dss_abt: Argobots Helpers
 */

#ifndef DAOS_SRV_ABT_H
#define DAOS_SRV_ABT_H

#define DABT_THREAD_JOIN(thread)								\
	do {											\
		int dabt_rc;									\
												\
		dabt_rc = ABT_thread_join(thread);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_thread_join: %d\n", dabt_rc);		\
	} while (0)

#define DABT_THREAD_FREE(thread)								\
	do {											\
		int dabt_rc;									\
												\
		dabt_rc = ABT_thread_free(thread);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_thread_free(%p): %d\n", thread, dabt_rc);\
	} while (0)

#define DABT_MUTEX_FREE(mutex)									\
	do {											\
		int dabt_rc;									\
												\
		dabt_rc = ABT_mutex_free(mutex);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_mutex_free(%p): %d\n", mutex, dabt_rc);	\
	} while (0)

#define DABT_COND_WAIT(cond, mutex)								\
	do {											\
		int dabt_rc;									\
												\
		dabt_rc = ABT_cond_wait(cond, mutex);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_cond_wait: %d\n", dabt_rc);		\
	} while (0)

#define DABT_COND_SIGNAL(cond)									\
	do {											\
		int dabt_rc;									\
												\
		dabt_rc = ABT_cond_signal(cond);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_cond_signal: %d\n", dabt_rc);		\
	} while (0)

#define DABT_COND_BROADCAST(cond)								\
	do {											\
		int dabt_rc;									\
												\
		dabt_rc = ABT_cond_broadcast(cond);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_cond_broadcast: %d\n", dabt_rc);		\
	} while (0)

#define DABT_EVENTUAL_SET(eventual, value, nbytes)						\
	do {											\
		int dabt_rc;									\
												\
		dabt_rc = ABT_eventual_set(eventual, value, nbytes);				\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_eventual_set(%p, %d): %d\n", value,	\
			  nbytes, dabt_rc);							\
	} while (0)

#define DABT_EVENTUAL_WAIT(eventual, value)							\
	do {											\
		int dabt_rc;									\
												\
		dabt_rc = ABT_eventual_wait(eventual, value);					\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_eventual_wait(%p): %d\n", value,		\
			  dabt_rc);								\
	} while (0)

#define DABT_EVENTUAL_FREE(eventual)								\
	do {											\
		int dabt_rc;									\
												\
		dabt_rc = ABT_eventual_free(eventual);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_eventual_free(%p): %d\n", eventual,	\
			  dabt_rc);								\
	} while (0)

#define DABT_FUTURE_SET(future, value)								\
	do {											\
		int dabt_rc;									\
												\
		dabt_rc = ABT_future_set(future, value);					\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_future_set(%p): %d\n", value, dabt_rc);	\
	} while (0)

#endif /* DAOS_SRV_ABT_H */
