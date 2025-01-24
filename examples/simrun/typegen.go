// Code generated by "core generate -add-types -add-funcs"; DO NOT EDIT.

package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/types"
	"cogentcore.org/lab/examples/baremetal"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensorcore"
)

var _ = types.AddType(&types.Type{Name: "main.FilterResults", IDName: "filter-results", Doc: "FilterResults specifies which results files to open.", Fields: []types.Field{{Name: "FileContains", Doc: "File name contains this string, e.g., \"_epc\" or \"_run\""}, {Name: "Ext", Doc: "extension of files, e.g., .tsv"}}})

var _ = types.AddType(&types.Type{Name: "main.SubmitParams", IDName: "submit-params", Doc: "SubmitParams specifies the parameters for submitting a job.", Fields: []types.Field{{Name: "Message", Doc: "Message describing the simulation:\nthis is key info for what is special about this job, like a github commit message"}, {Name: "Label", Doc: "Label is brief, unique label used for plots to label this job"}, {Name: "Args", Doc: "\targuments to pass on the command line.\n\n-nogui is already passed by default"}}})

var _ = types.AddType(&types.Type{Name: "main.JobParams", IDName: "job-params", Doc: "JobParams are parameters for running the job", Fields: []types.Field{{Name: "NRuns", Doc: "NRuns is the number of parallel runs; can also set to 1\nand run multiple runs per job using args."}, {Name: "Hours", Doc: "Hours is the max number of hours: slurm will terminate if longer,\nso be generous 2d = 48, 3d = 72, 4d = 96, 5d = 120, 6d = 144, 7d = 168"}, {Name: "Memory", Doc: "Memory per CPU in gigabytes"}, {Name: "Tasks", Doc: "Tasks is the number of mpi \"tasks\" (procs in MPI terminology)."}, {Name: "CPUsPerTask", Doc: "CPUsPerTask is the number of cpu cores (threads) per task."}, {Name: "TasksPerNode", Doc: "TasksPerNode is how to allocate tasks within compute nodes\ncpus_per_task * tasks_per_node <= total cores per node."}, {Name: "Qos", Doc: "Qos is the queue \"quality of service\" name."}}})

var _ = types.AddType(&types.Type{Name: "main.ServerParams", IDName: "server-params", Doc: "ServerParams are parameters for the server.", Fields: []types.Field{{Name: "Name", Doc: "Name is the name of current server using to run jobs;\ngets recorded with each job."}, {Name: "Root", Doc: "Root is the root path from user home dir on server.\nis auto-set to: filepath.Join(\"simdata\", Project, User)"}, {Name: "Slurm", Doc: "Slurm uses the slurm job manager. Otherwise uses a bare job manager."}}})

var _ = types.AddType(&types.Type{Name: "main.Configuration", IDName: "configuration", Doc: "Configuration holds all of the user-settable parameters", Fields: []types.Field{{Name: "DataRoot", Doc: "DataRoot is the path to the root of the data to browse."}, {Name: "StartDir", Doc: "StartDir is the starting directory, where the app was originally started."}, {Name: "User", Doc: "User id as in system login name (i.e., user@system)."}, {Name: "UserShort", Doc: "UserShort is the first 3 letters of User,\nfor naming jobs (auto-set from User)."}, {Name: "Project", Doc: "Project is the name of simulation project, lowercase\n(should be name of source dir)."}, {Name: "Package", Doc: "Package is the parent package: e.g., github.com/emer/axon/v2\nThis is used to update the go.mod, along with the Version."}, {Name: "Version", Doc: "Version is the current git version string, from git describe --tags."}, {Name: "Job", Doc: "Job has the parameters for job resources etc."}, {Name: "Server", Doc: "Server has server parameters."}, {Name: "FetchFiles", Doc: "FetchFiles is a glob expression for files to fetch from server,\nfor Fetch command. Is *.tsv by default."}, {Name: "ExcludeNodes", Doc: "ExcludeNodes are nodes to exclude from job, based on what is slow."}, {Name: "ExtraFiles", Doc: "ExtraFiles has extra files to upload with job submit, from same dir."}, {Name: "ExtraDirs", Doc: "ExtraDirs has subdirs with other files to upload with job submit\n(non-code -- see CodeDirs)."}, {Name: "CodeDirs", Doc: "CodeDirs has subdirs with code to upload with job submit;\ngo.mod auto-updated to use."}, {Name: "ExtraGoGet", Doc: "ExtraGoGet is an extra package to do \"go get\" with, for launching the job."}, {Name: "JobScript", Doc: "JobScript is a job script to use for running the simulation,\ninstead of the basic default, if non-empty.\nThis is written to the job.sbatch file. If it contains a $JOB_ARGS string\nthen that is replaced with the args entered during submission.\nIf using slurm, this switches to a simple direct sbatch submission instead\nof the default parallel job submission. All standard slurm job parameters\nare automatically inserted at the start of the file, so this script should\njust be the actual job running actions after that point."}, {Name: "SetupScript", Doc: "SetupScript contains optional lines of bash script code to insert at\nthe start of the job submission script, which is then followed by\nthe default script. For example, if a symbolic link to a large shared\nresource is needed, make that link here."}, {Name: "TimeFormat", Doc: "TimeFormat is the format for timestamps,\ndefaults to \"2006-01-02 15:04:05 MST\""}, {Name: "Filter", Doc: "Filter has parameters for filtering results."}, {Name: "Submit", Doc: "Submit has parameters for submitting jobs; set from last job run."}}})

var _ = types.AddType(&types.Type{Name: "main.Result", IDName: "result", Doc: "Result has info for one loaded result, as a table.Table", Fields: []types.Field{{Name: "JobID", Doc: "job id for results"}, {Name: "Label", Doc: "short label used as a legend in the plot"}, {Name: "Message", Doc: "description of job"}, {Name: "Args", Doc: "args used in running job"}, {Name: "Path", Doc: "path to data"}, {Name: "Table", Doc: "result data"}}})

var _ = types.AddType(&types.Type{Name: "main.SimRun", IDName: "sim-run", Doc: "SimRun manages the running and data analysis of results from simulations\nthat are run on remote server(s), within a Cogent Lab browser environment,\nwith the files as the left panel, and the Tabber as the right panel.", Methods: []types.Method{{Name: "FetchJobBare", Doc: "FetchJobBare downloads results files from bare metal server.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"jid", "force"}}, {Name: "EditConfig", Doc: "EditConfig edits the configuration", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Jobs", Doc: "Jobs updates the Jobs tab with a Table showing all the Jobs\nwith their meta data. Uses the dbmeta.toml data compiled from\nthe Status function.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "UpdateSims", Doc: "Jobs updates the Jobs tab with a Table showing all the Jobs\nwith their meta data. Uses the dbmeta.toml data compiled from\nthe Status function.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Queue", Doc: "Queue runs a queue query command on the server and shows the results.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Status", Doc: "Status gets updated job.* files from the server for any job that\ndoesn't have a Finalized or Fetched status.  It updates the\nstatus based on the server job status query, assigning a\nstatus of Finalized if job is done.  Updates the dbmeta.toml\ndata based on current job data.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Fetch", Doc: "Fetch retrieves all the .tsv data files from the server\nfor any jobs not already marked as Fetched.\nOperates on the jobs selected in the Jobs table,\nor on all jobs if none selected.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Cancel", Doc: "Cancel cancels the jobs selected in the Jobs table,\nwith a confirmation prompt.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Delete", Doc: "Delete deletes the selected Jobs, with a confirmation prompt.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Archive", Doc: "Archive moves the selected Jobs to the Archive directory,\nlocally, and deletes them from the server,\nfor results that are useful but not immediately relevant.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Results", Doc: "Results loads specific .tsv data files from the jobs selected\nin the Jobs table, into the Results table.  There are often\nmultiple result files per job, so this step is necessary to\nchoose which such files to select for plotting.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Diff", Doc: "Diff shows the differences between two selected jobs, or if only\none job is selected, between that job and the current source directory.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Plot", Doc: "Plot concatenates selected Results data files and generates a plot\nof the resulting data.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Reset", Doc: "Reset resets the Results table", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "Submit", Doc: "Submit submits a job to SLURM on the server, using an array\nstructure, with an outer startup job that calls the main array\njobs and a final cleanup job.  Creates a new job dir based on\nincrementing counter, synchronizing the job files.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}}, Embeds: []types.Field{{Name: "Frame"}, {Name: "Browser"}}, Fields: []types.Field{{Name: "Config", Doc: "Config holds all the configuration settings."}, {Name: "JobsTableView", Doc: "JobsTableView is the view of the jobs table."}, {Name: "JobsTable", Doc: "JobsTable is the jobs Table with one row per job."}, {Name: "ResultsTableView", Doc: "ResultsTableView has the results table."}, {Name: "ResultsList", Doc: "ResultsList is the list of result records."}, {Name: "BareMetal", Doc: "for now including directly -- will be rpc"}, {Name: "BareMetalActiveTable"}, {Name: "BareMetalDoneTable"}}})

// NewSimRun returns a new [SimRun] with the given optional parent:
// SimRun manages the running and data analysis of results from simulations
// that are run on remote server(s), within a Cogent Lab browser environment,
// with the files as the left panel, and the Tabber as the right panel.
func NewSimRun(parent ...tree.Node) *SimRun { return tree.New[SimRun](parent...) }

// SetConfig sets the [SimRun.Config]:
// Config holds all the configuration settings.
func (t *SimRun) SetConfig(v Configuration) *SimRun { t.Config = v; return t }

// SetJobsTableView sets the [SimRun.JobsTableView]:
// JobsTableView is the view of the jobs table.
func (t *SimRun) SetJobsTableView(v *tensorcore.Table) *SimRun { t.JobsTableView = v; return t }

// SetJobsTable sets the [SimRun.JobsTable]:
// JobsTable is the jobs Table with one row per job.
func (t *SimRun) SetJobsTable(v *table.Table) *SimRun { t.JobsTable = v; return t }

// SetResultsTableView sets the [SimRun.ResultsTableView]:
// ResultsTableView has the results table.
func (t *SimRun) SetResultsTableView(v *core.Table) *SimRun { t.ResultsTableView = v; return t }

// SetResultsList sets the [SimRun.ResultsList]:
// ResultsList is the list of result records.
func (t *SimRun) SetResultsList(v ...*Result) *SimRun { t.ResultsList = v; return t }

// SetBareMetal sets the [SimRun.BareMetal]:
// for now including directly -- will be rpc
func (t *SimRun) SetBareMetal(v *baremetal.Client) *SimRun { t.BareMetal = v; return t }

// SetBareMetalActiveTable sets the [SimRun.BareMetalActiveTable]
func (t *SimRun) SetBareMetalActiveTable(v *core.Table) *SimRun { t.BareMetalActiveTable = v; return t }

// SetBareMetalDoneTable sets the [SimRun.BareMetalDoneTable]
func (t *SimRun) SetBareMetalDoneTable(v *core.Table) *SimRun { t.BareMetalDoneTable = v; return t }

var _ = types.AddFunc(&types.Func{Name: "main.main", Doc: "important: must be run from an interactive terminal.\nWill quit immediately if not!"})

var _ = types.AddFunc(&types.Func{Name: "main.Interactive", Doc: "Interactive is the cli function that gets called by default at gui startup.", Args: []string{"c", "in"}, Returns: []string{"error"}})

var _ = types.AddFunc(&types.Func{Name: "main.NewSimRunWindow", Doc: "NewSimRunWindow returns a new SimRun window using given interpreter.\ndo RunWindow on resulting [core.Body] to open the window.", Args: []string{"in"}, Returns: []string{"Body", "SimRun"}})
