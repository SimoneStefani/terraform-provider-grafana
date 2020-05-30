package grafana

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	gapi "github.com/nytm/go-grafana-api"
)

func ResourceDataSource() *schema.Resource {
	return &schema.Resource{
		Create: resourceDataSourceCreate,
		Update: resourceDataSourceUpdate,
		Delete: resourceDataSourceDelete,
		Read:   resourceDataSourceRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"access_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "proxy",
			},

			"url": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"is_default": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"basic_auth_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"username": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "",
				Sensitive: true,
			},

			"database_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"json_data": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"client_email": {
							Type:     schema.TypeString,
							Required: true,
						},
						"default_project": {
							Type:     schema.TypeString,
							Required: true,
						},
						"token_uri": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"secure_json_data": {
				Type:      schema.TypeList,
				Optional:  true,
				Sensitive: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_key": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
	}
}

func resourceDataSourceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	dataSource, err := makeDataSource(d)
	if err != nil {
		return err
	}

	id, err := client.NewDataSource(dataSource)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(id, 10))

	return resourceDataSourceRead(d, meta)
}

func resourceDataSourceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	dataSource, err := makeDataSource(d)
	if err != nil {
		return err
	}

	return client.UpdateDataSource(dataSource)
}

func resourceDataSourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	idStr := d.Id()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid id: %#v", idStr)
	}

	dataSource, err := client.DataSource(id)
	if err != nil {
		if err.Error() == "404 Not Found" {
			log.Printf("[WARN] removing datasource %s from state because it no longer exists in grafana", d.Get("name").(string))
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("id", dataSource.Id)
	d.Set("org_id", dataSource.OrgId)
	d.Set("name", dataSource.Name)
	d.Set("type", dataSource.Type)
	d.Set("access_mode", dataSource.Access)
	d.Set("url", dataSource.URL)
	d.Set("is_default", dataSource.IsDefault)
	d.Set("basic_auth_enabled", dataSource.BasicAuth)
	d.Set("username", dataSource.User)
	d.Set("password", dataSource.Password)
	d.Set("database_name", dataSource.Database)

	return nil
}

func resourceDataSourceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	idStr := d.Id()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid id: %#v", idStr)
	}

	return client.DeleteDataSource(id)
}

func makeDataSource(d *schema.ResourceData) (*gapi.DataSource, error) {
	idStr := d.Id()
	var id int64
	var err error
	if idStr != "" {
		id, err = strconv.ParseInt(idStr, 10, 64)
	}

	return &gapi.DataSource{
		Id:             id,
		OrgId:          d.Get("org_id").(int64),
		Name:           d.Get("name").(string),
		Type:           d.Get("type").(string),
		Access:         d.Get("access_mode").(string),
		URL:            d.Get("url").(string),
		IsDefault:      d.Get("is_default").(bool),
		BasicAuth:      d.Get("basic_auth_enabled").(bool),
		User:           d.Get("username").(string),
		Password:       d.Get("password").(string),
		Database:       d.Get("database_name").(string),
		JSONData:       makeJSONData(d),
		SecureJSONData: makeSecureJSONData(d),
	}, err
}

func makeJSONData(d *schema.ResourceData) gapi.JSONData {
	return gapi.JSONData{
		AuthenticationType: d.Get("json_data.0.authentication_type").(string),
		ClientEmail:        d.Get("json_data.0.client_email").(string),
		DefaultProject:     d.Get("json_data.0.default_project").(string),
		TokenUri:           d.Get("json_data.0.token_uri").(string),
	}
}

func makeSecureJSONData(d *schema.ResourceData) gapi.SecureJSONData {
	return gapi.SecureJSONData{
		PrivateKey: d.Get("secure_json_data.0.private_key").(string),
	}
}
