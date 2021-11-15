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
		ABT_thread	dabt_thread = (thread);						\
		int		dabt_rc;							\
												\
		dabt_rc = ABT_thread_join(dabt_thread);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_thread_join: %d\n", dabt_rc);		\
	} while (0)

#define DABT_THREAD_FREE(thread)								\
	do {											\
		ABT_thread     *dabt_thread = (thread);						\
		int		dabt_rc;							\
												\
		dabt_rc = ABT_thread_free(dabt_thread);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_thread_free(%p): %d\n", dabt_thread,	\
			  dabt_rc);								\
	} while (0)

#define DABT_MUTEX_FREE(mutex)									\
	do {											\
		ABT_mutex      *dabt_mutex = (mutex);						\
		int		dabt_rc;							\
												\
		dabt_rc = ABT_mutex_free(dabt_mutex);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_mutex_free(%p): %d\n", dabt_mutex,	\
			  dabt_rc);								\
	} while (0)

#define DABT_COND_WAIT(cond, mutex)								\
	do {											\
		ABT_cond	dabt_cond = (cond);						\
		ABT_mutex	dabt_mutex = (mutex);						\
		int		dabt_rc;							\
												\
		dabt_rc = ABT_cond_wait(dabt_cond, dabt_mutex);					\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_cond_wait: %d\n", dabt_rc);		\
	} while (0)

#define DABT_COND_SIGNAL(cond)									\
	do {											\
		ABT_cond	dabt_cond = (cond);						\
		int		dabt_rc;							\
												\
		dabt_rc = ABT_cond_signal(dabt_cond);						\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_cond_signal: %d\n", dabt_rc);		\
	} while (0)

#define DABT_COND_BROADCAST(cond)								\
	do {											\
		ABT_cond	dabt_cond = (cond);						\
		int		dabt_rc;							\
												\
		dabt_rc = ABT_cond_broadcast(dabt_cond);					\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_cond_broadcast: %d\n", dabt_rc);		\
	} while (0)

#define DABT_EVENTUAL_SET(eventual, value, nbytes)						\
	do {											\
		ABT_eventual	dabt_eventual = (eventual);					\
		void	       *dabt_value = (value);						\
		int		dabt_nbytes = (nbytes);						\
		int		dabt_rc;							\
												\
		dabt_rc = ABT_eventual_set(dabt_eventual, dabt_value, dabt_nbytes);		\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_eventual_set(%p, %d): %d\n", dabt_value,	\
			  dabt_nbytes, dabt_rc);						\
	} while (0)

#define DABT_EVENTUAL_WAIT(eventual, value)							\
	do {											\
		ABT_eventual	dabt_eventual = (eventual);					\
		void	      **dabt_value = (value);						\
		int		dabt_rc;							\
												\
		dabt_rc = ABT_eventual_wait(dabt_eventual, dabt_value);				\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_eventual_wait(%p): %d\n", dabt_value,	\
			  dabt_rc);								\
	} while (0)

#define DABT_EVENTUAL_FREE(eventual)								\
	do {											\
		ABT_eventual   *dabt_eventual = (eventual);					\
		int		dabt_rc;							\
												\
		dabt_rc = ABT_eventual_free(dabt_eventual);					\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_eventual_free(%p): %d\n", dabt_eventual,	\
			  dabt_rc);								\
	} while (0)

#define DABT_FUTURE_SET(future, value)								\
	do {											\
		ABT_future	dabt_future = (future);						\
		void	       *dabt_value = (value);						\
		int		dabt_rc;							\
												\
		dabt_rc = ABT_future_set(dabt_future, dabt_value);				\
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_future_set(%p): %d\n", dabt_value,	\
			  dabt_rc);								\
	} while (0)

#endif /* DAOS_SRV_ABT_H */
