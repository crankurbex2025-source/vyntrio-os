package restore

import (
	"io"
	"os"
	"os/user"
	"strconv"
)

// Ownership applies fixed ownership and mode to restored targets.
type Ownership struct {
	StateUID  int
	StateGID  int
	ConfigUID int
	ConfigGID int
}

func (o Ownership) resolved() (Ownership, error) {
	out := o
	if out.StateUID == 0 && out.StateGID == 0 {
		uid, gid, err := lookupAccount(StateServiceAccount, StateServiceAccount)
		if err != nil {
			return Ownership{}, err
		}
		out.StateUID = uid
		out.StateGID = gid
	}
	if out.ConfigUID == 0 {
		out.ConfigUID = 0
	}
	if out.ConfigGID == 0 {
		_, gid, err := lookupAccount("", ConfigGroup)
		if err != nil {
			return Ownership{}, err
		}
		out.ConfigGID = gid
	}
	return out, nil
}

func lookupAccount(username, groupname string) (uid int, gid int, err error) {
	if username != "" {
		u, err := user.Lookup(username)
		if err != nil {
			return 0, 0, err
		}
		uid, err = strconv.Atoi(u.Uid)
		if err != nil {
			return 0, 0, err
		}
	}
	if groupname != "" {
		g, err := user.LookupGroup(groupname)
		if err != nil {
			return 0, 0, err
		}
		gid, err = strconv.Atoi(g.Gid)
		if err != nil {
			return 0, 0, err
		}
	}
	return uid, gid, nil
}

func applyOwnership(paths OwnershipTargets, ownership Ownership) error {
	own, err := ownership.resolved()
	if err != nil {
		return err
	}
	if err := os.Chown(paths.StateRoot, own.StateUID, own.StateGID); err != nil {
		return err
	}
	if err := os.Chmod(paths.StateRoot, stateDirMode); err != nil {
		return err
	}
	for _, path := range paths.StateFiles {
		if err := os.Chown(path, own.StateUID, own.StateGID); err != nil {
			return err
		}
		if err := os.Chmod(path, stateFileMode); err != nil {
			return err
		}
	}
	if paths.ConfigFile != "" {
		if err := os.Chown(paths.ConfigFile, own.ConfigUID, own.ConfigGID); err != nil {
			return err
		}
		if err := os.Chmod(paths.ConfigFile, configFileMode); err != nil {
			return err
		}
	}
	return nil
}

// OwnershipTargets lists host paths that require ownership repair.
type OwnershipTargets struct {
	StateRoot   string
	StateFiles  []string
	ConfigFile  string
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
