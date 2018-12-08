/**
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package realis

import (
	"strconv"

	"github.com/paypal/gorealis/gen-go/apache/aurora"
)

// Structure to collect all information pertaining to an Aurora job.
type AuroraJob struct {
	jobConfig *aurora.JobConfiguration
	resources map[string]*aurora.Resource
	portCount int
}

// Create a AuroraJob object with everything initialized.
func NewJob() *AuroraJob {

	jobKey := &aurora.JobKey{}

	// Task clientConfig
	taskConfig := &aurora.TaskConfig{
		Job:              jobKey,
		MesosFetcherUris: make([]*aurora.MesosFetcherURI, 0),
		Metadata:         make([]*aurora.Metadata, 0),
		Constraints:      make([]*aurora.Constraint, 0),
		// Container is a Union so one container field must be set. Set Mesos by default.
		Container: NewMesosContainer().Build(),
	}

	// AuroraJob clientConfig
	jobConfig := &aurora.JobConfiguration{
		Key:        jobKey,
		TaskConfig: taskConfig,
	}

	// Resources
	numCpus := &aurora.Resource{}
	ramMb := &aurora.Resource{}
	diskMb := &aurora.Resource{}

	numCpus.NumCpus = new(float64)
	ramMb.RamMb = new(int64)
	diskMb.DiskMb = new(int64)

	resources := make(map[string]*aurora.Resource)
	resources["cpu"] = numCpus
	resources["ram"] = ramMb
	resources["disk"] = diskMb

	taskConfig.Resources = []*aurora.Resource{numCpus, ramMb, diskMb}

	return &AuroraJob{
		jobConfig: jobConfig,
		resources: resources,
		portCount: 0,
	}
}

// Set AuroraJob Key environment.
func (j *AuroraJob) Environment(env string) *AuroraJob {
	j.jobConfig.Key.Environment = env
	return j
}

// Set AuroraJob Key Role.
func (j *AuroraJob) Role(role string) *AuroraJob {
	j.jobConfig.Key.Role = role

	// Will be deprecated
	identity := &aurora.Identity{User: role}
	j.jobConfig.Owner = identity
	j.jobConfig.TaskConfig.Owner = identity
	return j
}

// Set AuroraJob Key Name.
func (j *AuroraJob) Name(name string) *AuroraJob {
	j.jobConfig.Key.Name = name
	return j
}

// Set name of the executor that will the task will be configured to.
func (j *AuroraJob) ExecutorName(name string) *AuroraJob {

	if j.jobConfig.TaskConfig.ExecutorConfig == nil {
		j.jobConfig.TaskConfig.ExecutorConfig = aurora.NewExecutorConfig()
	}

	j.jobConfig.TaskConfig.ExecutorConfig.Name = name
	return j
}

// Will be included as part of entire task inside the scheduler that will be serialized.
func (j *AuroraJob) ExecutorData(data string) *AuroraJob {

	if j.jobConfig.TaskConfig.ExecutorConfig == nil {
		j.jobConfig.TaskConfig.ExecutorConfig = aurora.NewExecutorConfig()
	}

	j.jobConfig.TaskConfig.ExecutorConfig.Data = data
	return j
}

func (j *AuroraJob) CPU(cpus float64) *AuroraJob {
	*j.resources["cpu"].NumCpus = cpus

	return j
}

func (j *AuroraJob) RAM(ram int64) *AuroraJob {
	*j.resources["ram"].RamMb = ram

	return j
}

func (j *AuroraJob) Disk(disk int64) *AuroraJob {
	*j.resources["disk"].DiskMb = disk

	return j
}

func (j *AuroraJob) Tier(tier string) *AuroraJob {
	*j.jobConfig.TaskConfig.Tier = tier

	return j
}

// How many failures to tolerate before giving up.
func (j *AuroraJob) MaxFailure(maxFail int32) *AuroraJob {
	j.jobConfig.TaskConfig.MaxTaskFailures = maxFail
	return j
}

// How many instances of the job to run
func (j *AuroraJob) InstanceCount(instCount int32) *AuroraJob {
	j.jobConfig.InstanceCount = instCount
	return j
}

func (j *AuroraJob) CronSchedule(cron string) *AuroraJob {
	j.jobConfig.CronSchedule = &cron
	return j
}

func (j *AuroraJob) CronCollisionPolicy(policy aurora.CronCollisionPolicy) *AuroraJob {
	j.jobConfig.CronCollisionPolicy = policy
	return j
}

// How many instances of the job to run
func (j *AuroraJob) GetInstanceCount() int32 {
	return j.jobConfig.InstanceCount
}

// Restart the job's tasks if they fail
func (j *AuroraJob) IsService(isService bool) *AuroraJob {
	j.jobConfig.TaskConfig.IsService = isService
	return j
}

// Get the current job configurations key to use for some realis calls.
func (j *AuroraJob) JobKey() *aurora.JobKey {
	return j.jobConfig.Key
}

// Get the current job configurations key to use for some realis calls.
func (j *AuroraJob) JobConfig() *aurora.JobConfiguration {
	return j.jobConfig
}

func (j *AuroraJob) TaskConfig() *aurora.TaskConfig {
	return j.jobConfig.TaskConfig
}

// Add a list of URIs with the same extract and cache configuration. Scheduler must have
// --enable_mesos_fetcher flag enabled. Currently there is no duplicate detection.
func (j *AuroraJob) AddURIs(extract bool, cache bool, values ...string) *AuroraJob {
	for _, value := range values {
		j.jobConfig.TaskConfig.MesosFetcherUris = append(
			j.jobConfig.TaskConfig.MesosFetcherUris,
			&aurora.MesosFetcherURI{Value: value, Extract: &extract, Cache: &cache})
	}
	return j
}

// Adds a Mesos label to the job. Note that Aurora will add the
// prefix "org.apache.aurora.metadata." to the beginning of each key.
func (j *AuroraJob) AddLabel(key string, value string) *AuroraJob {
	j.jobConfig.TaskConfig.Metadata = append(j.jobConfig.TaskConfig.Metadata, &aurora.Metadata{Key: key, Value: value})
	return j
}

// Add a named port to the job configuration  These are random ports as it's
// not currently possible to request specific ports using Aurora.
func (j *AuroraJob) AddNamedPorts(names ...string) *AuroraJob {
	j.portCount += len(names)
	for _, name := range names {
		j.jobConfig.TaskConfig.Resources = append(j.jobConfig.TaskConfig.Resources, &aurora.Resource{NamedPort: &name})
	}

	return j
}

// Adds a request for a number of ports to the job configuration. The names chosen for these ports
// will be org.apache.aurora.port.X, where X is the current port count for the job configuration
// starting at 0. These are random ports as it's not currently possible to request
// specific ports using Aurora.
func (j *AuroraJob) AddPorts(num int) *AuroraJob {
	start := j.portCount
	j.portCount += num
	for i := start; i < j.portCount; i++ {
		portName := "org.apache.aurora.port." + strconv.Itoa(i)
		j.jobConfig.TaskConfig.Resources = append(j.jobConfig.TaskConfig.Resources, &aurora.Resource{NamedPort: &portName})
	}

	return j
}

// From Aurora Docs:
// Add a Value constraint
// name - Mesos slave attribute that the constraint is matched against.
// If negated = true , treat this as a 'not' - to avoid specific values.
// Values - list of values we look for in attribute name
func (j *AuroraJob) AddValueConstraint(name string, negated bool, values ...string) *AuroraJob {
	j.jobConfig.TaskConfig.Constraints = append(j.jobConfig.TaskConfig.Constraints,
		&aurora.Constraint{
			Name: name,
			Constraint: &aurora.TaskConstraint{
				Value: &aurora.ValueConstraint{
					Negated: negated,
					Values:  values,
				},
				Limit: nil,
			},
		})

	return j
}

// From Aurora Docs:
// A constraint that specifies the maximum number of active tasks on a host with
// a matching attribute that may be scheduled simultaneously.
func (j *AuroraJob) AddLimitConstraint(name string, limit int32) *AuroraJob {
	j.jobConfig.TaskConfig.Constraints = append(j.jobConfig.TaskConfig.Constraints,
		&aurora.Constraint{
			Name: name,
			Constraint: &aurora.TaskConstraint{
				Value: nil,
				Limit: &aurora.LimitConstraint{Limit: limit},
			},
		})

	return j
}

// From Aurora Docs:
// dedicated attribute. Aurora treats this specially, and only allows matching jobs
// to run on these machines, and will only schedule matching jobs on these machines.
// When a job is created, the scheduler requires that the $role component matches
// the role field in the job configuration, and will reject the job creation otherwise.
// A wildcard (*) may be used for the role portion of the dedicated attribute, which
// will allow any owner to elect for a job to run on the host(s)
func (j *AuroraJob) AddDedicatedConstraint(role, name string) *AuroraJob {
	j.AddValueConstraint("dedicated", false, role+"/"+name)

	return j
}

// Set a container to run for the job configuration to run.
func (j *AuroraJob) Container(container Container) *AuroraJob {
	j.jobConfig.TaskConfig.Container = container.Build()

	return j
}
