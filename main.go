package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

func main() {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal("failed to create session,", err)
	}

	svc := route53.New(sess)

	// Get env variables
	domain := os.Getenv("DOMAIN")
	recordType := os.Getenv("TYPE")
	records := strings.Split(os.Getenv("RECORDS"), ",")

	rrs, err := getResourceRecordSets(svc)
	if err != nil {
		log.Fatal(err)
	}
	frrs, err := filterResourceRecordSets(rrs, domain, recordType)
	if err != nil {
		log.Fatal(err)
	}
	registeredIP, err := getFirstRecordSet(frrs)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Get("https://checkip.amazonaws.com/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	homeIP := strings.Trim(string(body), " \n")

	if homeIP == registeredIP {
		log.Print(fmt.Sprintf("Registered ip address matches home ip address: %s", homeIP))
	} else {
		log.Print("Update Route53 A record")
		createARecord(svc, records, homeIP)
	}

}

func getResourceRecordSets(svc *route53.Route53) (*route53.ListResourceRecordSetsOutput, error) {
	listParams := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(os.Getenv("HOSTED_ZONE_ID")),
	}
	return svc.ListResourceRecordSets(listParams)
}

func filterResourceRecordSets(rrs *route53.ListResourceRecordSetsOutput, domain string, recordType string) (*route53.ResourceRecordSet, error) {
	for _, v := range rrs.ResourceRecordSets {
		currentName := *v.Name
		currentType := *v.Type
		if currentName == domain && currentType == recordType {
			return v, nil
		}
	}
	return nil, fmt.Errorf("No matching ResourceRecordSets for Name: %s, Type: %s", domain, recordType)
}

func getFirstRecordSet(r *route53.ResourceRecordSet) (ipAddress string, err error) {
	if len(r.ResourceRecords) == 0 {
		errStr := fmt.Sprintf("ResourceRecordSet does not have any resource records.")
		err = errors.New(errStr)
		return ipAddress, err
	}
	return *r.ResourceRecords[0].Value, nil
}

func isIPAddress(str string) bool {
	re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	return re.MatchString(str)
}

func createARecord(svc *route53.Route53, records []string, ipAddress string) {
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
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			return
		}

		// Pretty-print the response data.
		fmt.Println("Change Response:")
		fmt.Println(resp)
	}
}
