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
		D_ASSERTF(dabt_rc == ABT_SUCCESS, "ABT_eventual_free: %d\n", dabt_rc);		\
	} while (0)

#endif /* DAOS_SRV_ABT_H */
