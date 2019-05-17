package exoscale

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

var supportedRecordTypes = []string{
	"A", "AAAA", "ALIAS", "CAA", "CNAME",
	"HINFO", "MX", "NAPTR", "NS", "POOL",
	"SPF", "SRV", "SSHFP", "TXT", "URL",
}

func resourceExoscaleDomainRecordIDString(d resourceIDStringer) string {
	return resourceIDString(d, "exoscale_domain_record")
}

func domainRecordResource() *schema.Resource {
	return &schema.Resource{
		Create: createRecord,
		Read:   readRecord,
		Exists: existsRecord,
		Update: updateRecord,
		Delete: deleteRecord,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"record_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(supportedRecordTypes, true),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"content": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"prio": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Read:   schema.DefaultTimeout(defaultTimeout),
			Update: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},
	}
}

func createRecord(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] %s: beginning create", resourceExoscaleDomainRecordIDString(d))

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	client := GetDNSClient(meta)

	record, err := client.CreateRecord(ctx, d.Get("domain").(string), egoscale.DNSRecord{
		Name:       d.Get("name").(string),
		Content:    d.Get("content").(string),
		RecordType: d.Get("record_type").(string),
		TTL:        d.Get("ttl").(int),
		Prio:       d.Get("prio").(int),
	})

	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(record.ID, 10))

	log.Printf("[DEBUG] %s: create finished successfully", resourceExoscaleDomainRecordIDString(d))

	return readRecord(d, meta)
}

func existsRecord(d *schema.ResourceData, meta interface{}) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	client := GetDNSClient(meta)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	domain := d.Get("domain").(string)

	if domain != "" {
		record, err := client.GetRecord(ctx, domain, id)
		if err != nil {
			if _, ok := err.(*egoscale.DNSErrorResponse); !ok {
				return false, err
			}

			return true, err
		}

		return record != nil, nil
	}

	domains, err := client.GetDomains(ctx)
	if err != nil {
		return false, err
	}

	for _, domain := range domains {
		record, err := client.GetRecord(ctx, domain.Name, id)
		if err != nil {
			if _, ok := err.(*egoscale.DNSErrorResponse); !ok {
				return false, err
			}

			return true, err
		}

		if record != nil {
			return true, nil
		}
	}

	return false, nil
}

func readRecord(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] %s: beginning read", resourceExoscaleDomainRecordIDString(d))

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	client := GetDNSClient(meta)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	domain := d.Get("domain").(string)

	// TODO: when is it not the case? Isn't the `domain` attribute supposed to be mandatory?
	if domain != "" {
		record, err := client.GetRecord(ctx, domain, id)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] %s: read finished successfully", resourceExoscaleDomainRecordIDString(d))

		return applyRecord(d, *record)
	}

	domains, err := client.GetDomains(ctx)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		record, err := client.GetRecord(ctx, domain.Name, id)
		if err != nil {
			return err
		}

		if record != nil {
			if err := d.Set("domain", domain.Name); err != nil {
				return err
			}

			log.Printf("[DEBUG] %s: read finished successfully", resourceExoscaleDomainRecordIDString(d))

			return applyRecord(d, *record)
		}
	}

	return fmt.Errorf("domain record %s not found", d.Id())
}

func updateRecord(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] %s: beginning update", resourceExoscaleDomainRecordIDString(d))

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	client := GetDNSClient(meta)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	record, err := client.UpdateRecord(ctx, d.Get("domain").(string), egoscale.UpdateDNSRecord{
		ID:      id,
		Name:    d.Get("name").(string),
		Content: d.Get("content").(string),
		TTL:     d.Get("ttl").(int),
		Prio:    d.Get("prio").(int),
	})

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] %s: update finished successfully", resourceExoscaleDomainRecordIDString(d))

	return applyRecord(d, *record)
}

func deleteRecord(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] %s: beginning delete", resourceExoscaleDomainRecordIDString(d))

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	client := GetDNSClient(meta)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	err := client.DeleteRecord(ctx, d.Get("domain").(string), id)
	if err != nil { // FIXME: isn't it supposed to be the opposite? (if err == nil)
		d.SetId("")
	}

	log.Printf("[DEBUG] %s: delete finished successfully", resourceExoscaleDomainRecordIDString(d))

	return err
}

func applyRecord(d *schema.ResourceData, record egoscale.DNSRecord) error {
	d.SetId(strconv.FormatInt(record.ID, 10))
	if err := d.Set("name", record.Name); err != nil {
		return err
	}
	if err := d.Set("content", record.Content); err != nil {
		return err
	}
	if err := d.Set("record_type", record.RecordType); err != nil {
		return err
	}
	if err := d.Set("ttl", record.TTL); err != nil {
		return err
	}
	if err := d.Set("prio", record.Prio); err != nil {
		return err
	}

	domain := d.Get("domain").(string)
	if record.Name != "" {
		domain = fmt.Sprintf("%s.%s", record.Name, domain)
	}

	if err := d.Set("hostname", domain); err != nil {
		return nil
	}

	return nil
}
