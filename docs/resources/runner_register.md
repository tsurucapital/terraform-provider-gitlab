# gitlab\_runner\_register

This resource allows you to create and manage GitLab runner registrations. For
further information on registering GitLab runners, consult the [GitLab
documentation](https://docs.gitlab.com/ee/api/runners.html#register-a-new-runner).

## Example Usage

```hcl
resource "gitlab_runner_register" "example" {
   registration_token = "fooxpba-RbAzquux1234567"
   tags = ["terraform"]
}
```

## Argument Reference

The following arguments are supported:

* `registration_token` - (Required, string) Token used to [register the
  runner](https://docs.gitlab.com/runner/register/). It can be [obtained through
  GitLab](https://docs.gitlab.com/ee/ci/runners/README.html).

* `description` - (Optional, string) Runner's description.

* `name` - (Optional) Runner name. Currently [not
  supported](https://github.com/xanzy/go-gitlab/issues/1003). You might want to
  use `description` instead.

* `active` - (Optional, boolean) Whether the runner is active.

* `locked` - (Optional, boolean) Whether the runner should be locked for current
  project.

* `run_untagged` - (Optional, boolean) Whether the runner should handle untagged
  jobs.

* `tags` - (Optional, list of strings) List of runner’s tags.

* `access_level` - (Optional, string) The access_level of the runner;
  `not_protected` or `ref_protected`.

* `maximum_timeout` - (Optional, integer) Maximum timeout set when this runner
  handles the job.

## Attributes Reference

The following additional attributes are exported:

* `id` - Runner ID that uniquely identifies the runner within the GitLab
  install.

* `token` - GitLab authentication token used to be used by the runner.

* `version` - GitLab runner version being used such as `13.6.0`.

* `revision` - GitLab runner revision being used such as `8fa89735`.

* `platform` - Platform the runner is running on such as `linux`.

* `architecture` - Archictecture of the machine the runner is running on such as
  `amd64`.

* `ip_address` - IP address of the machine the GitLab runner is running on.

* `is_shared` - Indicates if this is a shared runner.

* `contacted_at` - Last time the GitLab runner machine made contact with the
  GitLab instance.

* `online` - Indicates if the GitLab runner is online.

* `status` - Status of the runner such as `paused` or `online`.


#### Arguments

Only `version` is displayed in the Admin area of the UI.

## Importing pre-registered runners

You can import a runner state using `terraform import <resource> <id>`.  The
`id` can be whatever the [get runner's details API][https://docs.gitlab.com/ee/api/runners.html#get-runners-details] takes for
its `:id` value, so for example:

    terraform import gitlab_project.example 123
