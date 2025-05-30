#!/bin/bash

cwd=$(dirname "$0")
DAOS_DIR=${DAOS_DIR:-$(cd "$cwd/../../.." && echo "$PWD")}
# shellcheck disable=SC1091
source "${DAOS_DIR}/.build_vars.sh"
BTR=${SL_BUILD_DIR}/src/common/tests/btree
VCMD=""
BAT_NUM=${BAT_NUM:-"20000"}
if [ "$USE_VALGRIND" = "memcheck" ]; then
    BAT_NUM="200"
    VCMD="valgrind --leak-check=full --show-reachable=yes --error-limit=no \
          --suppressions=${VALGRIND_SUPP} --error-exitcode=42 --xml=yes \
          --xml-file=unit-test-btree-%p.memcheck.xml"
elif [ "$USE_VALGRIND" = "pmemcheck" ]; then
    VCMD="valgrind --tool=pmemcheck"
fi

ORDER=${ORDER:-11}
DDEBUG=${DDEBUG:-0}


KEYS=${KEYS:-"3,6,5,7,2,1,4"}
RECORDS=${RECORDS:-"7:loaded,3:that,5:dice,2:knows,4:the,6:are,1:Everybody"}

function print_help()
{
    cat << EOF
Usage: btree.sh [OPTIONS]
    Options:
        -s [num]  Run with num keys
        dyn       Run with dynamic root
        ukey      Use integer keys
        emb       Use embedded value
        perf      Run performance tests
        direct    Use direct string key
EOF
    exit 1
}

PERF=""
UINT=""
DYN=""
test_conf_pre=""
while [ $# -gt 0 ]; do
    case "$1" in
    -s)
        shift
        BAT_NUM=$1
        if [ "$BAT_NUM" -ne "$BAT_NUM" ]; then
            echo "Bad argument to -s option.  Must be numeric"
            print_help
        fi
        shift
        test_conf_pre="${test_conf_pre} keys=${BAT_NUM}"
        ;;
    dyn)
        DYN="d"
        shift
        test_conf_pre="${test_conf_pre} dyn"
        ;;
    perf)
        shift
        PERF="on"
        test_conf_pre="${test_conf_pre} perf"
        ;;
    ukey)
        shift
        UINT="+"
        test_conf_pre="${test_conf_pre} ukey"
        ;;
    emb)
        shift
        # reuse this flag since they are mutually exclusive
        UINT="%"
        test_conf_pre="${test_conf_pre} ukey"
        ;;
    direct)
        BTR=${SL_BUILD_DIR}/src/common/tests/btree_direct
        KEYS=${KEYS:-"delta,lambda,kappa,omega,beta,alpha,epsilon"}
        RECORDS=${RECORDS:-"omega:loaded,delta:that,kappa:dice,beta:knows,epsilon:the,lambda:are,alpha:Everybody"}
        shift
        test_conf_pre="${test_conf_pre} direct"
        ;;
    *)
        echo "Unknown option $1"
        print_help
        ;;
    esac
done

set -x
set -e

gen_test_conf_string()
{
        name=""
        [ -n "$1" ] && name="${name} inplace"
        [ -n "$2" ] && name="${name} pmem"
        echo "$name"
}

run_test()
{
    printf "\nOptions: IPL='%s' IDIR='%s' PMEM='%s'\n" "$IPL" "$IDIR" "$PMEM"
    test_conf=$(gen_test_conf_string "${IPL}" "${PMEM}")

    if [ -z ${PERF} ]; then

        echo "B+tree functional test..."
        DAOS_DEBUG="$DDEBUG"                        \
        eval "${VCMD}" "$BTR"                       \
        --start-test "'btree functional ${test_conf_pre} ${test_conf} iterate=${IDIR}'" \
        -R"${DYN}" -M"${PMEM}" -C "${UINT}${IPL}o:$ORDER" \
        -c                                          \
        -o                                          \
        -u "$RECORDS"                               \
        -i "$IDIR"                                  \
        -q                                          \
        -f "$KEYS"                                  \
        -d "$KEYS"                                  \
        -u "$RECORDS"                               \
        -f "$KEYS"                                  \
        -r "$KEYS"                                  \
        -q                                          \
        -u "$RECORDS"                               \
        -q                                          \
        -i "$IDIR:3"                                \
        -D

        echo "B+tree batch operations test..."
        eval "${VCMD}" "$BTR" \
        --start-test "'btree batch operations ${test_conf_pre} ${test_conf}'" \
        -R"${DYN}" -M"${PMEM}" -C "${UINT}${IPL}o:$ORDER" \
        -c                                          \
        -o                                          \
        -b "$BAT_NUM"                               \
        -D

        echo "B+tree drain test..."
        eval "${VCMD}" "$BTR" \
        --start-test "'btree drain ${test_conf_pre} ${test_conf}'" \
        -R"${DYN}" -M"${PMEM}" -C "${UINT}${IPL}o:$ORDER" \
        -e -D

    else
        echo "B+tree performance test..."
        eval "${VCMD}" "$BTR" \
        --start-test "'btree performance ${test_conf_pre} ${test_conf}'" \
        -R"${DYN}" -M"${PMEM}" -C "${UINT}${IPL}o:$ORDER" \
        -p "$BAT_NUM"                               \
        -D
    fi
}

for IPL in "i," ""; do
    for IDIR in "f" "b"; do
        for PMEM in "p" ""; do
            run_test
        done
    done
done
