#include <stdarg.h>
#include <stddef.h>
#include <setjmp.h>
#include <cmocka.h>

#include <daos_srv/rdb.h>
#include "rdb_internal.h"

static void
test_rdb_encode_decode_iov(void **state)
{
	d_iov_t	iov;
	uint8_t	buf[13];
	size_t	len = sizeof(buf) / 2;
	uint8_t	buf_out[sizeof(buf) + 8];
	d_iov_t	iov_out;
	size_t	rc;

	iov.iov_buf = buf;
	iov.iov_buf_len = sizeof(buf);
	iov.iov_len = len;
	memset(buf, 'a', len);

	rc = rdb_encode_iov(&iov, NULL /* buf */);
	assert_int_equal(rc, 4 /* size_head */ + len + 4 /* size_tail */);
	assert_true(rc <= sizeof(buf_out));

	rc = rdb_encode_iov(&iov, buf_out);
	assert_int_equal(rc, 4 /* size_head */ + len + 4 /* size_tail */);

	iov_out.iov_buf = NULL;
	iov_out.iov_buf_len = 0;
	iov_out.iov_len = 0;

	rc = rdb_decode_iov(buf_out, sizeof(buf_out), &iov_out);
	assert_int_equal(rc, 4 /* size_head */ + len + 4 /* size_tail */);
	assert_non_null(iov_out.iov_buf);
	assert_int_equal(iov_out.iov_buf_len, len);
	assert_int_equal(iov_out.iov_len, len);
	assert_memory_equal(iov_out.iov_buf, iov.iov_buf, len);
}

int
main(void)
{
	const struct CMUnitTest tests[] = {
		cmocka_unit_test(test_rdb_encode_decode_iov)
	};

	return cmocka_run_group_tests(tests, NULL, NULL);
}
