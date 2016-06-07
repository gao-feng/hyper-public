package daemon

import (
	"fmt"
	"github.com/golang/glog"
	"strings"
)

// tag is nil means resize tty of container
func (daemon *Daemon) TtyResize(podId, tag string, h, w int) error {
	var (
		container string
		execId    string
		vmid      string
		err       error
		pod       *Pod
		ok        bool
	)

	if strings.Contains(podId, "pod-") {
		vmid, err = daemon.GetVmByPodId(podId)
		if err != nil {
			return err
		}
		pod, ok = daemon.PodList.Get(podId)
		if !ok {
			return fmt.Errorf("can not find pod %s", podId)
		}

		if tag != "" {
			// want to resize tty of container?
			for _, c := range pod.ctnInfo {
				if _, ok := c.ClientTag[tag]; ok {
					container = c.Id
					break
				}
			}

			pod.RLock()
			// want to resize tty of exec?
			if exec, ok := pod.execList[tag]; ok {
				execId = exec.ExecId
				container = exec.Container
			}
			pod.RUnlock()
		} else {
			ttyContainers := pod.ctnInfo
			if pod.spec.Type == "service-discovery" {
				ttyContainers = pod.ctnInfo[1:]
			}

			if len(ttyContainers) >= 1 {
				container = ttyContainers[0].Id
			}
		}

		if container == "" {
			return fmt.Errorf("can not find container for pod %s, tag %s", podId, tag)
		}
	} else if strings.Contains(podId, "vm-") {
		// Doesn't support resize vm process's tty
		return fmt.Errorf("doesn't support id %s", podId)
	} else {
		container = podId
		pod, _, err = daemon.GetPodByContainerIdOrName(container)
		if err != nil {
			return err
		}

		vmid, err = daemon.GetVmByPodId(pod.Id)
		if err != nil {
			return err
		}

		pod.RLock()
		// want to resize tty of exec
		if exec, ok := pod.execList[tag]; ok {
			execId = exec.ExecId
		}
		pod.RUnlock()
	}

	vm, ok := daemon.VmList[vmid]
	if !ok {
		return fmt.Errorf("vm %s doesn't exist!", vmid)
	}

	err = vm.Tty(container, execId, h, w)
	if err != nil {
		return err
	}

	glog.V(1).Infof("Success to resize the tty!")
	return nil
}
