// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: BUSL-1.1

package feasible

import (
	log "github.com/hashicorp/go-hclog"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hashicorp/nomad/nomad/structs"
	sstructs "github.com/hashicorp/nomad/scheduler/structs"
)

// DependencyChecker is a FeasibilityChecker which returns nodes that match a
// given set of dependencies. This is used to filter on job, task group, and task
// dependencies.
type DependencyChecker struct {
	log     log.Logger
	state   sstructs.State
	job     *structs.Job
	metrics *structs.AllocMetric
	ready   bool
}

// NewDependencyChecker creates a DependencyChecker for a set of dependencies
func NewDependencyChecker(c Context, dependencies []*structs.Dependency) *DependencyChecker {
	return &DependencyChecker{
		log:     c.Logger().Named("dependency_checker"),
		state:   c.State(),
		ready:   true,
		metrics: c.Metrics(),
		job:     c.Plan().Job,
	}
}

func (dc *DependencyChecker) SetDependencies(dependencies []*structs.Dependency) {
	ws := memdb.NewWatchSet()
	for _, dep := range dependencies {
		job, err := dc.state.JobByID(ws, dc.job.Namespace, dep.Job)
		if err != nil {
			dc.log.Error("error looking up dependency job", "dependency_job_id", dep.Job, "error", err)
		}

		dc.ready = dc.ready && (job != nil && job.Status == dep.Output)
	}
}

func (dc *DependencyChecker) Feasible(option *structs.Node) bool {
	if !dc.ready {
		dc.metrics.FilterNode(option, "dependecy_not_ready")
		return false
	}

	return true
}
