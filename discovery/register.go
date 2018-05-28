package discovery

import (
	"context"

	"github.com/Bilibili/discovery/errors"
	"github.com/Bilibili/discovery/model"
	log "github.com/golang/glog"
)

// Register a new instance.
func (d *Discovery) Register(c context.Context, ins *model.Instance, arg *model.ArgRegister) {
	d.registry.Register(ins, arg.LatestTimestamp)
	if !arg.Replication {
		d.nodes.Replicate(c, model.Register, ins, arg.Zone != d.c.Zone)
	}
}

// Renew marks the given instance of the given app name as renewed, and also marks whether it originated from replication.
func (d *Discovery) Renew(c context.Context, arg *model.ArgRenew) (i *model.Instance, err error) {
	i, ok := d.registry.Renew(arg)
	if !ok {
		log.Errorf("renew appid(%s) hostname(%s) zone(%s) env(%s) error", arg.AppID, arg.Hostname, arg.Zone, arg.Env)
		return
	}
	if !arg.Replication {
		d.nodes.Replicate(c, model.Renew, i, arg.Zone != d.c.Zone)
		return
	}
	if arg.DirtyTimestamp > i.DirtyTimestamp {
		err = errors.NothingFound
		return
	} else if arg.DirtyTimestamp < i.DirtyTimestamp {
		err = errors.Conflict
	}
	return
}

// Cancel cancels the registration of an instance.
func (d *Discovery) Cancel(c context.Context, arg *model.ArgCancel) (err error) {
	i, ok := d.registry.Cancel(arg)
	if !ok {
		err = errors.NothingFound
		log.Errorf("cancel appid(%s) hostname(%s) error", arg.AppID, arg.Hostname)
		return
	}
	if !arg.Replication {
		d.nodes.Replicate(c, model.Cancel, i, arg.Zone != d.c.Zone)
	}
	return
}

// FetchAll fetch all instances of all the department.
func (d *Discovery) FetchAll(c context.Context) (im map[string][]*model.Instance) {
	return d.registry.FetchAll()
}

// Fetch fetch all instances by appid.
func (d *Discovery) Fetch(c context.Context, arg *model.ArgFetch) (info *model.InstanceInfo, err error) {
	return d.registry.Fetch(arg.Zone, arg.Env, arg.AppID, 0, arg.Status)
}

// Fetchs fetch multi app by appids.
func (d *Discovery) Fetchs(c context.Context, arg *model.ArgFetchs) (is map[string]*model.InstanceInfo, err error) {
	is = make(map[string]*model.InstanceInfo, len(arg.AppID))
	for _, appid := range arg.AppID {
		i, err := d.registry.Fetch(arg.Zone, arg.Env, appid, 0, arg.Status)
		if err != nil {
			log.Errorf("Fetchs fetch appid(%v) err", err)
			continue
		}
		is[appid] = i
	}
	return
}

// Polls hangs request and then write instances when that has changes, or return NotModified.
func (d *Discovery) Polls(c context.Context, arg *model.ArgPolls) (ch chan map[string]*model.InstanceInfo, new bool, err error) {
	return d.registry.Polls(arg)
}

// DelConns delete conn of host in appid
func (d *Discovery) DelConns(arg *model.ArgPolls) {
	d.registry.DelConns(arg)
}

// Nodes get all nodes of discovery.
func (d *Discovery) Nodes(c context.Context) (nsi []*model.Node) {
	return d.nodes.Nodes()
}

// PutChan put chan into pool.
func (d *Discovery) PutChan(ch chan map[string]*model.InstanceInfo) {
	d.registry.PutChan(ch)
}