//
// (C) Copyright 2021-2024 Intel Corporation.
// (C) Copyright 2025 Google LLC
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
//

package main

/*
#include "util.h"

static void
free_daos_alloc(void *ptr)
{
        // Use the macro to free memory allocated
        // by DAOS macros in order to keep NLT happy.
        D_FREE(ptr);
}
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/pkg/errors"

	"github.com/daos-stack/daos/src/control/lib/daos"
)

func dfsError(rc C.int) error {
	if rc == 0 {
		return nil
	}

	strErr := C.strerror(rc)
	return errors.New(fmt.Sprintf("errno %d (%s)", rc, C.GoString(strErr)))
}

type fsCmd struct {
	Check          fsCheckCmd          `command:"check" description:"Run the DFS Checker on the container"`
	FixSB          fsFixSBCmd          `command:"fix-sb" description:"Recreate / Fix the SB on the container"`
	FixRoot        fsFixRootCmd        `command:"fix-root" description:"Relink root object in the container"`
	FixEntry       fsFixEntryCmd       `command:"fix-entry" description:"Fix Entries in case the type or chunk size were corrupted"`
	Copy           fsCopyCmd           `command:"copy" description:"copy to and from a POSIX filesystem"`
	SetAttr        fsSetAttrCmd        `command:"set-attr" description:"set default object class and/or chunk size"`
	GetAttr        fsGetAttrCmd        `command:"get-attr" description:"get default object class and chunk size"`
	ResetAttr      fsResetAttrCmd      `command:"reset-attr" description:"reset default creation attributes on a directory"`
	ResetChunkSize fsResetChunkSizeCmd `command:"reset-chunk-size" description:"reset default creation chunk size on a directory"`
	ResetObjClass  fsResetOclassCmd    `command:"reset-oclass" description:"reset default creation object class on a directory"`
	DfuseQuery     fsDfuseQueryCmd     `command:"query" description:"Query dfuse for memory usage"`
	DfuseEvict     fsDfuseEvictCmd     `command:"evict" description:"Evict object from dfuse"`
	Scan           fsScanCmd           `command:"scan" description:"Scan POSIX container and report statistics"`
	Chmod          fsChmodCmd          `command:"chmod" description:"Change file mode bits"`
}

type fsCopyCmd struct {
	daosCmd

	Source      string `long:"src" short:"s" description:"copy source" required:"1"`
	Dest        string `long:"dst" short:"d" description:"copy destination" required:"1"`
	Preserve    string `long:"preserve-props" short:"m" description:"preserve container properties, requires HDF5 library" required:"0"`
	IgnoreUnsup bool   `long:"ignore-unsupported" description:"ignore unsupported filesystem features when copying to DFS" required:"0"`
}

func (cmd *fsCopyCmd) Execute(_ []string) error {
	ap, deallocCmdArgs, err := allocCmdArgs(cmd.Logger)
	if err != nil {
		return err
	}
	defer deallocCmdArgs()

	ap.src = C.CString(cmd.Source)
	defer freeString(ap.src)
	ap.dst = C.CString(cmd.Dest)
	defer freeString(ap.dst)
	if cmd.Preserve != "" {
		ap.preserve_props = C.CString(cmd.Preserve)
		defer freeString(ap.preserve_props)
	}
	ap.ignore_unsup = C.bool(cmd.IgnoreUnsup)

	ap.fs_op = C.FS_COPY
	rc := C.fs_copy_hdlr(ap)
	if err := daosError(rc); err != nil {
		return errors.Wrapf(err, "failed to copy %s -> %s", cmd.Source, cmd.Dest)
	}

	if cmd.JSONOutputEnabled() {
		type CopyStats struct {
			NumDirs  uint64 `json:"num_dirs"`
			NumFiles uint64 `json:"num_files"`
			NumLinks uint64 `json:"num_links"`
		}

		return cmd.OutputJSON(struct {
			SourcePool string    `json:"src_pool"`
			SourceCont string    `json:"src_cont"`
			DestPool   string    `json:"dst_pool"`
			DestCont   string    `json:"dst_cont"`
			CopyStats  CopyStats `json:"copy_stats"`
		}{
			C.GoString(&ap.dm_args.src_pool[0]),
			C.GoString(&ap.dm_args.src_cont[0]),
			C.GoString(&ap.dm_args.dst_pool[0]),
			C.GoString(&ap.dm_args.dst_cont[0]),
			CopyStats{
				uint64(ap.fs_copy_stats.num_dirs),
				uint64(ap.fs_copy_stats.num_files),
				uint64(ap.fs_copy_stats.num_links),
			},
		}, nil)
	}

	fsType := "DAOS"
	if ap.fs_copy_posix {
		fsType = "POSIX"
	}
	// Compat with old-style output
	cmd.Infof("Successfully copied to %s: %s", fsType, cmd.Dest)
	cmd.Infof("    Directories: %d", ap.fs_copy_stats.num_dirs)
	cmd.Infof("    Files:       %d", ap.fs_copy_stats.num_files)
	cmd.Infof("    Links:       %d", ap.fs_copy_stats.num_links)

	if ap.fs_copy_stats.num_chmod_enotsup > 0 {
		return errors.Errorf("Copy completed successfully, but %d files had unsupported mode bits that could not be applied. Run with --ignore-unsupported to suppress this warning.", ap.fs_copy_stats.num_chmod_enotsup)
	}

	return nil
}

type fsAttrCmd struct {
	existingContainerCmd

	DfsPath   string `long:"dfs-path" short:"H" description:"DFS path relative to root of container (required when using pool and container instead of --path)"`
	DfsPrefix string `long:"dfs-prefix" short:"I" description:"Optional prefix path to the root of the DFS container when using pool and container"`
}

func setupFSAttrCmd(cmd *fsAttrCmd) (*C.struct_cmd_args_s, func(), error) {
	ap, deallocCmdArgs, err := allocCmdArgs(cmd.Logger)
	if err != nil {
		return nil, nil, err
	}

	if cmd.DfsPath == "" && cmd.Path == "" {
		deallocCmdArgs()
		return nil, nil, errors.New("If not using --path, --dfs-path must be specified along with pool/container IDs")
	}
	if cmd.DfsPath != "" {
		if cmd.Path != "" {
			deallocCmdArgs()
			return nil, nil, errors.New("Cannot use both --dfs-path and --path")
		}
		ap.dfs_path = C.CString(cmd.DfsPath)
	} else {
		if cmd.Path == "" {
			deallocCmdArgs()
			return nil, nil, errors.New("--dfs-path is required if not using --path")
		}
	}
	if cmd.DfsPrefix != "" {
		if cmd.Path != "" {
			deallocCmdArgs()
			return nil, nil, errors.New("Cannot use both --dfs-prefix and --path")
		}
		ap.dfs_prefix = C.CString(cmd.DfsPrefix)
	}

	return ap, deallocCmdArgs, nil
}

func fsOpString(op uint32) string {
	switch op {
	case C.FS_SET_ATTR:
		return "set-attr"
	case C.FS_RESET_ATTR:
		return "reset-attr"
	case C.FS_RESET_CHUNK_SIZE:
		return "reset-chunk-size"
	case C.FS_RESET_OCLASS:
		return "reset-oclass"
	case C.FS_GET_ATTR:
		return "get-attr"
	}
	return "unknown operation"
}

func fsModifyAttr(cmd *fsAttrCmd, op uint32, updateAP func(*C.struct_cmd_args_s)) error {
	ap, deallocCmdArgs, err := setupFSAttrCmd(cmd)
	if err != nil {
		return err
	}
	defer deallocCmdArgs()

	ap.fs_op = op
	if updateAP != nil {
		updateAP(ap)
	}

	cleanup, err := cmd.resolveAndOpen(daos.ContainerOpenFlagReadWrite, ap)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := dfsError(C.fs_dfs_hdlr(ap)); err != nil {
		return errors.Wrapf(err, "%s failed", fsOpString(op))
	}

	return nil
}

type fsSetAttrCmd struct {
	fsAttrCmd

	ChunkSize   ChunkSizeFlag `long:"chunk-size" short:"z" description:"container chunk size"`
	ObjectClass ObjClassFlag  `long:"oclass" short:"o" description:"default object class"`
}

func (cmd *fsSetAttrCmd) Execute(_ []string) error {
	return fsModifyAttr(&cmd.fsAttrCmd, C.FS_SET_ATTR, func(ap *C.struct_cmd_args_s) {
		if cmd.ObjectClass.Set {
			ap.oclass = cmd.ObjectClass.Class
		}
		if cmd.ChunkSize.Set {
			ap.chunk_size = cmd.ChunkSize.Size
		}
	})
}

type fsResetAttrCmd struct {
	fsAttrCmd
}

func (cmd *fsResetAttrCmd) Execute(_ []string) error {
	return fsModifyAttr(&cmd.fsAttrCmd, C.FS_RESET_ATTR, nil)
}

type fsResetChunkSizeCmd struct {
	fsAttrCmd
}

func (cmd *fsResetChunkSizeCmd) Execute(_ []string) error {
	return fsModifyAttr(&cmd.fsAttrCmd, C.FS_RESET_CHUNK_SIZE, nil)
}

type fsResetOclassCmd struct {
	fsAttrCmd
}

func (cmd *fsResetOclassCmd) Execute(_ []string) error {
	return fsModifyAttr(&cmd.fsAttrCmd, C.FS_RESET_OCLASS, nil)
}

type fsGetAttrCmd struct {
	fsAttrCmd
}

func (cmd *fsGetAttrCmd) Execute(_ []string) error {
	ap, deallocCmdArgs, err := setupFSAttrCmd(&cmd.fsAttrCmd)
	if err != nil {
		return err
	}
	defer deallocCmdArgs()

	ap.fs_op = C.FS_GET_ATTR

	cleanup, err := cmd.resolveAndOpen(daos.ContainerOpenFlagReadOnly, ap)
	if err != nil {
		return err
	}
	defer cleanup()

	var attrs C.dfs_obj_info_t
	var cmode C.mode_t
	if err := dfsError(C.fs_dfs_get_attr_hdlr(ap, &attrs, &cmode)); err != nil {
		return errors.Wrapf(err, "%s failed", fsOpString((ap.fs_op)))
	}

	var oclassName [16]C.char
	C.daos_oclass_id2name(attrs.doi_oclass_id, &oclassName[0])

	var diroclassName [16]C.char
	var fileoclassName [16]C.char
	var oid C.daos_obj_id_t = attrs.doi_oid
	var oidStr string = fmt.Sprintf("%d.%d", oid.hi, oid.lo)
	if C.mode_is_dir(cmode) {
		C.daos_oclass_id2name(attrs.doi_dir_oclass_id, &diroclassName[0])
		C.daos_oclass_id2name(attrs.doi_file_oclass_id, &fileoclassName[0])
	}

	if cmd.JSONOutputEnabled() {
		if C.mode_is_dir(cmode) {
			jsonAttrs := struct {
				ObjAttr struct {
					OID      string `json:"oid"`
					ObjClass string `json:"oclass"`
				} `json:"object"`
				DirAttr struct {
					DirObjClass  string `json:"dir_oclass"`
					FileObjClass string `json:"file_oclass"`
					ChunkSize    uint64 `json:"chunk_size"`
				} `json:"directory"`
			}{
				ObjAttr: struct {
					OID      string `json:"oid"`
					ObjClass string `json:"oclass"`
				}{
					OID:      oidStr,
					ObjClass: C.GoString(&oclassName[0]),
				},
				DirAttr: struct {
					DirObjClass  string `json:"dir_oclass"`
					FileObjClass string `json:"file_oclass"`
					ChunkSize    uint64 `json:"chunk_size"`
				}{
					FileObjClass: C.GoString(&diroclassName[0]),
					DirObjClass:  C.GoString(&fileoclassName[0]),
					ChunkSize:    uint64(attrs.doi_chunk_size),
				},
			}
			return cmd.OutputJSON(jsonAttrs, nil)
		} else {
			jsonAttrs := &struct {
				OID       string `json:"oid"`
				ObjClass  string `json:"oclass"`
				ChunkSize uint64 `json:"chunk_size"`
			}{
				OID:       oidStr,
				ObjClass:  C.GoString(&oclassName[0]),
				ChunkSize: uint64(attrs.doi_chunk_size),
			}
			return cmd.OutputJSON(jsonAttrs, nil)
		}
	}

	cmd.Infof("OID = %s", oidStr)
	cmd.Infof("Object Class = %s", C.GoString(&oclassName[0]))
	if C.mode_is_dir(cmode) {
		cmd.Infof("Directory Creation Object Class = %s", C.GoString(&diroclassName[0]))
		cmd.Infof("File Creation Object Class = %s", C.GoString(&fileoclassName[0]))
		cmd.Infof("File Creation Chunk Size = %d", attrs.doi_chunk_size)
	} else {
		cmd.Infof("Object Chunk Size = %d", attrs.doi_chunk_size)
	}
	return nil
}

type fsCheckCmd struct {
	existingContainerCmd

	FsckFlags FsCheckFlag `long:"flags" short:"f" description:"comma-separated flags: print, remove, relink, verify, evict"`
	DirName   string      `long:"dir-name" short:"n" description:"directory name under lost+found to store leaked oids (a timestamp dir would be created if this is not specified)"`
}

func (cmd *fsCheckCmd) Execute(_ []string) error {
	ap, deallocCmdArgs, err := allocCmdArgs(cmd.Logger)
	if err != nil {
		return err
	}
	defer deallocCmdArgs()

	if err := cmd.resolveContainer(ap); err != nil {
		return err
	}

	cleanupPool, err := cmd.connectPool(C.DAOS_PC_RW, ap)
	if err != nil {
		return err
	}
	defer cleanupPool()

	var dirName *C.char
	if cmd.DirName != "" {
		dirName = C.CString(cmd.DirName)
		defer freeString(dirName)
	}

	rc := C.dfs_cont_check(cmd.cPoolHandle, &ap.cont_str[0], cmd.FsckFlags.Flags, dirName)
	if err := dfsError(rc); err != nil {
		return errors.Wrapf(err, "failed fs check")
	}
	return nil
}

type fsFixEntryCmd struct {
	fsAttrCmd

	ChunkSize ChunkSizeFlag `long:"chunk-size" short:"z" description:"actual file chunk size used when creating the file"`
	Type      bool          `long:"type" short:"t" description:"check the object of the entry and fix the entry type accordingly"`
}

func (cmd *fsFixEntryCmd) Execute(_ []string) error {
	ap, deallocCmdArgs, err := setupFSAttrCmd(&cmd.fsAttrCmd)
	if err != nil {
		return err
	}
	defer deallocCmdArgs()

	ap.fs_op = C.FS_CHECK
	cleanup, err := cmd.resolveAndOpen(daos.ContainerOpenFlagExclusive, ap)
	if err != nil {
		return errors.Wrapf(err, "failed fs fix-entry")
	}
	defer cleanup()

	ap.chunk_size = 0
	if cmd.ChunkSize.Set {
		ap.chunk_size = cmd.ChunkSize.Size
	}

	if err := dfsError(C.fs_fix_entry_hdlr(ap, C.bool(cmd.Type))); err != nil {
		return errors.Wrapf(err, "failed filesystem fix-entry")
	}

	return nil
}

type fsFixSBCmd struct {
	existingContainerCmd

	Mode            ConsModeFlag  `long:"mode" short:"M" description:"DFS consistency mode"`
	ChunkSize       ChunkSizeFlag `long:"chunk-size" short:"z" description:"container chunk size"`
	ObjectClass     ObjClassFlag  `long:"oclass" short:"o" description:"default object class"`
	DirObjectClass  ObjClassFlag  `long:"dir-oclass" short:"D" description:"default directory object class"`
	FileObjectClass ObjClassFlag  `long:"file-oclass" short:"F" description:"default file object class"`
	CHints          string        `long:"hints" short:"H" description:"container hints"`
}

func (cmd *fsFixSBCmd) Execute(_ []string) error {
	ap, deallocCmdArgs, err := allocCmdArgs(cmd.Logger)
	if err != nil {
		return err
	}
	defer deallocCmdArgs()

	ap.fs_op = C.FS_CHECK
	cleanup, err := cmd.resolveAndOpen(daos.ContainerOpenFlagExclusive, ap)
	if err != nil {
		return err
	}
	defer cleanup()

	if cmd.ChunkSize.Set {
		ap.chunk_size = cmd.ChunkSize.Size
	}
	if cmd.ObjectClass.Set {
		ap.oclass = cmd.ObjectClass.Class
	}
	if cmd.DirObjectClass.Set {
		ap.dir_oclass = cmd.DirObjectClass.Class
	}
	if cmd.FileObjectClass.Set {
		ap.file_oclass = cmd.FileObjectClass.Class
	}
	if cmd.Mode.Set {
		ap.mode = cmd.Mode.Mode
	}
	if cmd.CHints != "" {
		ap.hints = C.CString(cmd.CHints)
		defer freeString(ap.hints)
	}

	if err := dfsError(C.fs_recreate_sb_hdlr(ap)); err != nil {
		return errors.Wrapf(err, "Recreate SB failed")
	}

	return nil
}

type fsFixRootCmd struct {
	existingContainerCmd
}

func (cmd *fsFixRootCmd) Execute(_ []string) error {
	ap, deallocCmdArgs, err := allocCmdArgs(cmd.Logger)
	if err != nil {
		return err
	}
	defer deallocCmdArgs()

	ap.fs_op = C.FS_CHECK
	cleanup, err := cmd.resolveAndOpen(daos.ContainerOpenFlagExclusive, ap)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := dfsError(C.fs_relink_root_hdlr(ap)); err != nil {
		return errors.Wrapf(err, "Relink Root failed")
	}

	return nil
}

type fsDfuseQueryCmd struct {
	daosCmd

	Ino uint64 `long:"inode" description:"inode number to query"`

	Args struct {
		Path string `positional-arg-name:"path" description:"DFuse path to query" required:"1"`
	} `positional-args:"yes"`
}

func (cmd *fsDfuseQueryCmd) Execute(_ []string) error {
	ap, deallocCmdArgs, err := allocCmdArgs(cmd.Logger)
	if err != nil {
		return err
	}

	ap.path = C.CString(cmd.Args.Path)
	defer freeString(ap.path)
	defer deallocCmdArgs()

	if cmd.Ino != 0 {
		ap.dfuse_mem.ino = C.ulong(cmd.Ino)
	}

	rc := C.dfuse_cont_query(ap)
	if err := daosError(rc); err != nil {
		return errors.Wrapf(err, "failed to query %s", cmd.Args.Path)
	}

	defer C.free_daos_alloc(unsafe.Pointer(ap.dfuse_stat))

	if cmd.JSONOutputEnabled() {

		ds_map := make(map[string]uint64)

		ds_slice := unsafe.Slice(ap.dfuse_stat, ap.dfuse_mem.stat_count)

		for _, v := range ds_slice {
			if v.value != 0 {
				ds_map[strings.ToLower(C.GoString(&v.name[0]))] = uint64(v.value)
			}
		}

		if cmd.Ino == 0 {
			jsonAttrs := &struct {
				NumInodes      uint64            `json:"inodes"`
				NumFileHandles uint64            `json:"open_files"`
				NumPools       uint64            `json:"pools"`
				NumContainers  uint64            `json:"containers"`
				Data           map[string]uint64 `json:"statistics"`
			}{
				NumInodes:      uint64(ap.dfuse_mem.inode_count),
				NumFileHandles: uint64(ap.dfuse_mem.fh_count),
				NumPools:       uint64(ap.dfuse_mem.pool_count),
				NumContainers:  uint64(ap.dfuse_mem.container_count),
				Data:           ds_map,
			}
			return cmd.OutputJSON(jsonAttrs, nil)
		} else {
			jsonAttrs := &struct {
				NumInodes      uint64 `json:"inodes"`
				NumFileHandles uint64 `json:"open_files"`
				NumPools       uint64 `json:"pools"`
				NumContainers  uint64 `json:"containers"`
				Found          bool   `json:"resident"`
			}{
				NumInodes:      uint64(ap.dfuse_mem.inode_count),
				NumFileHandles: uint64(ap.dfuse_mem.fh_count),
				NumPools:       uint64(ap.dfuse_mem.pool_count),
				NumContainers:  uint64(ap.dfuse_mem.container_count),
				Found:          bool(ap.dfuse_mem.found),
			}
			return cmd.OutputJSON(jsonAttrs, nil)
		}
	}

	cmd.Infof("DFuse descriptor usage.")
	cmd.Infof("      Pools: %d", ap.dfuse_mem.pool_count)
	cmd.Infof(" Containers: %d", ap.dfuse_mem.container_count)
	cmd.Infof("     Inodes: %d", ap.dfuse_mem.inode_count)
	cmd.Infof(" Open files: %d", ap.dfuse_mem.fh_count)
	if cmd.Ino != 0 {
		if ap.dfuse_mem.found {
			cmd.Infof(" Inode %d resident", cmd.Ino)
		} else {
			cmd.Infof(" Inode %d not resident", cmd.Ino)
		}
	}

	return nil
}

type fsDfuseEvictCmd struct {
	daosCmd

	Args struct {
		Path string `positional-arg-name:"path" description:"Path to evict from dfuse" required:"1"`
	} `positional-args:"yes"`
}

func (cmd *fsDfuseEvictCmd) Execute(_ []string) error {
	ap, deallocCmdArgs, err := allocCmdArgs(cmd.Logger)
	if err != nil {
		return err
	}

	ap.path = C.CString(cmd.Args.Path)
	defer freeString(ap.path)
	defer deallocCmdArgs()

	rc := C.dfuse_evict(ap)
	if err := daosError(rc); err != nil {
		return errors.Wrapf(err, "failed to evict %s", cmd.Args.Path)
	}

	if cmd.JSONOutputEnabled() {
		jsonAttrs := &struct {
			NumInodes      uint64 `json:"inodes"`
			NumFileHandles uint64 `json:"open_files"`
			NumPools       uint64 `json:"pools"`
			NumContainers  uint64 `json:"containers"`
			Inode          uint64 `json:"inode,omitempty"`
		}{
			NumInodes:      uint64(ap.dfuse_mem.inode_count),
			NumFileHandles: uint64(ap.dfuse_mem.fh_count),
			NumPools:       uint64(ap.dfuse_mem.pool_count),
			NumContainers:  uint64(ap.dfuse_mem.container_count),
			Inode:          uint64(ap.dfuse_mem.ino),
		}
		return cmd.OutputJSON(jsonAttrs, nil)
	}

	cmd.Infof("DFuse descriptor usage.")
	cmd.Infof(" Evicted inode: %d", ap.dfuse_mem.ino)
	cmd.Infof("         Pools: %d", ap.dfuse_mem.pool_count)
	cmd.Infof("    Containers: %d", ap.dfuse_mem.container_count)
	cmd.Infof("        Inodes: %d", ap.dfuse_mem.inode_count)
	cmd.Infof("    Open files: %d", ap.dfuse_mem.fh_count)

	return nil
}

type fsScanCmd struct {
	existingContainerCmd

	DirName string `long:"dir-name" short:"n" description:"subdirectory path to scan. Start from container root if not specified."`
}

func (cmd *fsScanCmd) Execute(_ []string) error {
	ap, deallocCmdArgs, err := allocCmdArgs(cmd.Logger)
	if err != nil {
		return err
	}
	defer deallocCmdArgs()

	if err := cmd.resolveContainer(ap); err != nil {
		return err
	}

	cleanupPool, err := cmd.connectPool(C.DAOS_PC_RW, ap)
	if err != nil {
		return err
	}
	defer cleanupPool()

	var dirName *C.char
	if cmd.DirName != "" {
		dirName = C.CString(cmd.DirName)
		defer freeString(dirName)
	}

	rc := C.dfs_cont_scan(cmd.cPoolHandle, &ap.cont_str[0], 0, dirName)
	if err := dfsError(rc); err != nil {
		return errors.Wrapf(err, "failed to scan")
	}
	return nil
}

type fsChmodCmd struct {
	fsAttrCmd

	ModeBits ModeBitsFlag `long:"mode" short:"m" description:"file mode in octal, e.g. 0755" required:"1"`
}

func (cmd *fsChmodCmd) Execute(_ []string) error {
	ap, deallocCmdArgs, err := setupFSAttrCmd(&cmd.fsAttrCmd)
	if err != nil {
		return err
	}
	defer deallocCmdArgs()

	ap.object_mode = cmd.ModeBits.Mode

	cleanup, err := cmd.resolveAndOpen(daos.ContainerOpenFlagReadWrite, ap)
	if err != nil {
		return errors.Wrapf(err, "failed to connect")
	}
	defer cleanup()

	if err := dfsError(C.fs_chmod_hdlr(ap)); err != nil {
		return errors.Wrapf(err, "chmod failed")
	}

	return nil
}
