#!/bin/bash
#
#  Copyright 2020-2022 Intel Corporation.
#  Copyright 2025 Hewlett Packard Enterprise Development LP
#
#  SPDX-License-Identifier: BSD-2-Clause-Patent
#

post_provision_config_nodes() {
    # should we port this to Ubuntu or just consider $CONFIG_POWER_ONLY dead?
    #if $CONFIG_POWER_ONLY; then
    #    rm -f /etc/yum.repos.d/*.hpdd.intel.com_job_daos-stack_job_*_job_*.repo
    #    yum -y erase fio fuse ior-hpc mpich-autoload               \
    #                 ompi argobots cart daos daos-client dpdk      \
    #                 fuse-libs libisa-l libpmemobj mercury mpich   \
    #                 pmix protobuf-c spdk libfabric libpmem        \
    #                 munge-libs munge slurm                        \
    #                 slurm-example-configs slurmctld slurm-slurmmd
    #fi
    codename=$(lsb_release -s -c)
    echo "$codename"
    if [ -n "$INST_REPOS" ]; then
        for repo in $INST_REPOS; do
            branch="master"
            build_number="lastSuccessfulBuild"
            if [[ $repo = *@* ]]; then
                branch="${repo#*@}"
                repo="${repo%@*}"
                if [[ $branch = *:* ]]; then
                    build_number="${branch#*:}"
                    branch="${branch%:*}"
                fi
            fi
            echo "deb [trusted=yes] ${JENKINS_URL}job/daos-stack/job/${repo}/job/${branch//\//%252F}/${build_number}/artifact/artifacts/ubuntu20.04 ./" >> /etc/apt/sources.list
        done
    fi
    apt-get update
    if [ -n "$INST_RPMS" ]; then
        # shellcheck disable=SC2086
        if ! apt-get -y remove $INST_RPMS; then
            rc=${PIPESTATUS[0]}
            if [ $rc -ne 100 ]; then
                echo "Error $rc removing $INST_RPMS"
                return $rc
            fi
        fi
    fi

    apt-get -y install lsb-core

    # shellcheck disable=2086
    if [ -n "$INST_RPMS" ] &&
       ! apt-get -y install $INST_RPMS; then
        rc=${PIPESTATUS[0]}
        for file in /etc/apt/sources.list{,.d/*.list}; do
            echo "---- $file ----"
            cat "$file"
        done
        return "$rc"
    fi

    # change the default shell to bash -- we write a lot of bash
    chsh -s /bin/bash

    return 0
}
