package gitlab

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/xanzy/go-gitlab"
)

func resourceGitlabRunnerRegister() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabRunnerCreate,
		Read:   resourceGitlabRunnerRead,
		Update: resourceGitlabRunnerUpdate,
		Delete: resourceGitlabRunnerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGitlabRunnerRegisterImporter,
		},

		Schema: map[string]*schema.Schema{
			"registration_token": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revision": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"architecture": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"locked": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"run_untagged": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"tags": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"access_level": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "not_protected",

				ValidateFunc: validation.StringInSlice([]string{"not_protected", "ref_protected"}, false),
			},
			"maximum_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			// Computed read/stuff.
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"contacted_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"online": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceGitlabRunnerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	options := &gitlab.RegisterNewRunnerOptions{
		Token: gitlab.String(d.Get("registration_token").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		options.Description = gitlab.String(v.(string))
	}

	if _, ok := d.GetOk("name"); ok {
		return fmt.Errorf("Setting 'name' attribute is currently not supported due to https://github.com/xanzy/go-gitlab/issues/1003")
	}

	// Send the booleans that we have defaults for even if they aren't set: this
	// avoids bugs like https://gitlab.com/gitlab-org/gitlab/-/issues/208749 at
	// a very minor cost of sightly larger request.
	options.Active = gitlab.Bool(d.Get("active").(bool))
	options.Locked = gitlab.Bool(d.Get("locked").(bool))
	options.RunUntagged = gitlab.Bool(d.Get("run_untagged").(bool))

	if v, ok := d.GetOk("tags"); ok {
		options.TagList = *stringSetToStringSlice(v.(*schema.Set))
	}

	if v, ok := d.GetOk("maximum_timeout"); ok {
		options.MaximumTimeout = gitlab.Int(v.(int))
	}

	log.Printf("[DEBUG] create gitlab registered runner %q", *options.Token)
	runner, _, err := client.Runners.RegisterNewRunner(options)
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%d", runner.ID))
	d.Set("token", runner.Token)

	return resourceGitlabRunnerRead(d, meta)
}

func resourceGitlabRunnerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] read gitlab runner %s", d.Id())
	runner, resp, err := client.Runners.GetRunnerDetails(d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			log.Printf("[DEBUG] gitlab registered runner %s not found so removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] setting gitlab runner %#v from %#v", d, runner)
	d.SetId(fmt.Sprintf("%d", runner.ID))
	d.Set("active", runner.Active)
	d.Set("description", runner.Description)
	d.Set("ip_address", runner.IPAddress)
	d.Set("is_shared", runner.IsShared)
	d.Set("contacted_at", runner.ContactedAt)
	d.Set("online", runner.Online)
	d.Set("status", runner.Status)
	d.Set("projects", runner.Projects)
	// From GitLab 13.0 onwards, the token is no longer returned when listing
	// runner details: we only see it on the initial creation. The GitLab API
	// returns an empty string here in this case. We only set the token
	// parameter if there is some non-empty value.
	if runner.Token != "" {
		d.Set("token", runner.Token)
	}
	d.Set("tags", runner.TagList)
	d.Set("locked", runner.Locked)
	d.Set("access_level", runner.AccessLevel)
	d.Set("maximum_timeout", runner.MaximumTimeout)
	d.Set("groups", runner.Groups)
	d.Set("name", runner.Name)
	d.Set("version", runner.Version)
	d.Set("revision", runner.Revision)
	d.Set("platform", runner.Platform)
	d.Set("architecture", runner.Architecture)
	d.Set("run_untagged", runner.RunUntagged)
	log.Printf("[DEBUG] After setting we have %#v", d)
	return nil
}

func resourceGitlabRunnerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	options := &gitlab.UpdateRunnerDetailsOptions{}

	if d.HasChange("description") {
		options.Description = gitlab.String(d.Get("description").(string))
	}
	if d.HasChange("active") {
		options.Active = gitlab.Bool(d.Get("active").(bool))
	}
	if d.HasChange("tags") {
		options.TagList = *stringSetToStringSlice(d.Get("tags").(*schema.Set))
	}
	if d.HasChange("run_untagged") {
		options.RunUntagged = gitlab.Bool(d.Get("run_untagged").(bool))
	}
	if d.HasChange("locked") {
		options.Locked = gitlab.Bool(d.Get("locked").(bool))
	}
	if d.HasChange("access_level") {
		options.AccessLevel = gitlab.String(d.Get("access_level").(string))
	}
	if d.HasChange("maximum_timeout") {
		options.MaximumTimeout = gitlab.Int(d.Get("maximum_timeout").(int))
	}

	log.Printf("[DEBUG] update gitlab registered runner %s", d.Id())
	_, _, err := client.Runners.UpdateRunnerDetails(d.Id(), options)
	if err != nil {
		return err
	}
	return resourceGitlabRunnerRead(d, meta)
}

func resourceGitlabRunnerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] Delete gitlab registered runner %s", d.Id())
	_, err := client.Runners.RemoveRunner(d.Id())
	return err
}

func resourceGitlabRunnerRegisterImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := resourceGitlabRunnerRead(d, meta)
	return []*schema.ResourceData{d}, err
}
