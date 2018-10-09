package kubeless

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/kubeless/kubeless/pkg/functions"
	"io/ioutil"
	"net/http"
	"log"
	"os"
	"regexp"
	"strings"
)

func handler(event functions.Event, context functions.Context) (string, error) {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal("failed to create session,", err)
	}

	svc := route53.New(sess)

	// Get env variables
	domain := string("swhurl.com.")
	recordType := string("A")
	records := strings.Split(string("swhurl.com,*.swhurl.com"), ",")
	
	rrs, err := getResourceRecordSets(svc)
	if (err != nil) {
		log.Fatal(err)
	}
	frrs, err := filterResourceRecordSets(rrs, domain, recordType)
	if (err != nil) {
	  log.Fatal(err)
	}
  registeredIp, err := getFirstRecordSet(frrs)
	if (err != nil) {
	  log.Fatal(err)
	}

	resp, err := http.Get("https://checkip.amazonaws.com/")
	if (err != nil) {
	  log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	homeIp := strings.Trim(string(body), " \n")
	
	if (homeIp == registeredIp) {
	  log.Print(fmt.Sprintf("Registered ip address matches home ip address: %s", homeIp))
	} else {
	  log.Print("Update Route53 A record")
		response, err := createARecord(svc, records, homeIp)
		if (err != nil) {
      log.Fatal(err)
		}
	}
	return response, nil
	
}

func getResourceRecordSets(svc *route53.Route53) (*route53.ListResourceRecordSetsOutput, error) {
	listParams := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String("Z3NVP9T7260ZB8"),
	}
	return svc.ListResourceRecordSets(listParams)
}

func filterResourceRecordSets(rrs *route53.ListResourceRecordSetsOutput, domain string, recordType string) (*route53.ResourceRecordSet, error) {
	for _, v := range rrs.ResourceRecordSets {
		currentName := *v.Name
		currentType := *v.Type
		if (currentName == domain && currentType == recordType) {
			return v, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("No matching ResourceRecordSets for Name: %s, Type: %s", domain, recordType))
}

func getFirstRecordSet(r *route53.ResourceRecordSet) (ipAddress string, err error) {
	if (len(r.ResourceRecords) == 0) {
		errStr := fmt.Sprintf("ResourceRecordSet does not have any resource records.")
		err = errors.New(errStr)	
		return ipAddress, err
	} else {
		return *r.ResourceRecords[0].Value, nil
	}
}

func isIpAddress(str string) (bool) {
  re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
  return re.MatchString(str)
}

func createARecord(svc *route53.Route53, records []string, ipAddress string) (string, error) {
	response := ""
	for _, v := range records {
		params := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{ 
						Action: aws.String("UPSERT"),
						ResourceRecordSet: &route53.ResourceRecordSet{
							Name: aws.String(v),   
							Type: aws.String("A"),
							ResourceRecords: []*route53.ResourceRecord{
								{
									Value: aws.String(ipAddress),
								},
							},
							TTL:           aws.Int64(300),
							Weight:        aws.Int64(100),
							SetIdentifier: aws.String("home ip address"),
						},
					},
				},
				Comment: aws.String("Dynamic update"),
			},
			HostedZoneId: aws.String(os.Getenv("HOSTED_ZONE_ID")),
		}
		resp, err := svc.ChangeResourceRecordSets(params)

		if err != nil {
			return "", err
		}

		// Pretty-print the response data.
		fmt.Println("Change Response:")
		fmt.Println(resp)
		response += fmt.Sprintln(resp)
	}
	return response, nil
}

