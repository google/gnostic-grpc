Command to find OpenAPI 3.0 descriptions inside [APIs-guru/openapi-directory](https://github.com/APIs-guru/openapi-directory):
    
     find . -name "openapi.yaml" -exec gnostic --grpc-out=. {} \;
     
 Note:
 gnostic reading error: Gnostic throws an error while converting the .yaml to Proto Bufs. We see that as long as gnostic
 converts the OpenAPI description, the plugin gnostic-grpc works (with the improved surface model; see PR).

| API                                       | Status:                                       | Description |
| ------------------------------------------|:---------------------------------------------:|:---------------------------------------------:|
| ./APIs/namsor.com/2.0.5/openapi.yaml                          | OK                        ||
| ./APIs/bunq.com/1.0/openapi.yaml                              | gnostic reading error         | 
| ./APIs/ip2location.com/geolocation/1.0/openapi.yaml           | OK                        | 
| ./APIs/linode.com/4.0.15/openapi.yaml                         | gnostic reading error         | 
| ./APIs/geodatasource.com/1.0/openapi.yaml                     | OK                        | 
| ./APIs/ipinfodb.com/1.0/openapi.yaml                          | OK                        | 
| ./APIs/patrowl.local/1.0.0/openapi.yaml                       | OK                        | 
| ./APIs/cpy.re/peertube/1.3.1/openapi.yaml                     | gnostic reading error         | 
| ./APIs/twitter.com/labs/1.1/openapi.yaml                      | gnostic reading error     | OK after removing line 1051 inside yaml file | 
| ./APIs/api2pdf.com/1.0.0/openapi.yaml                         | OK                        | 
| ./APIs/exchangerate-api.com/4/openapi.yaml                    | OK                        | 
| ./APIs/bclaws.ca/bclaws/1.0.0/openapi.yaml                    | OK                        | 
| ./APIs/getthedata.com/bng2latlong/1.0/openapi.yaml            | OK                        | 
| ./APIs/gov.bc.ca/jobposting/1.0.0/openapi.yaml                | OK                        | 
| ./APIs/gov.bc.ca/bcgnws/3.x.x/openapi.yaml                    | OK                        | 
| ./APIs/gov.bc.ca/news/1.0/openapi.yaml                        | OK                        | 
| ./APIs/gov.bc.ca/gwells/v1/openapi.yaml                       | OK                        | 
| ./APIs/gov.bc.ca/open511/1.0.0/openapi.yaml                   | OK                        | 
| ./APIs/gov.bc.ca/bcdc/3.0.1/openapi.yaml                      | OK                        | 
| ./APIs/gov.bc.ca/geocoder/2.0.0/openapi.yaml                  | OK                        | 
| ./APIs/gov.bc.ca/geomark/4.1.2/openapi.yaml                   | OK                        | 
| ./APIs/gov.bc.ca/router/2.0.0/openapi.yaml                    | OK                        | 
| ./APIs/box.com/2.0/openapi.yaml                               | OK                        | | 
| ./APIs/departureboard.io/1.0/openapi.yaml                     | OK                        | 
| ./APIs/openlinksw.com/osdb/1.0.0/openapi.yaml                 | OK                        | 
| ./APIs/bbc.com/1.0.0/openapi.yaml                             | OK                        | | 
| ./APIs/whatsapp.local/1.0/openapi.yaml                        | OK                        | 
| ./APIs/vimeo.com/3.4/openapi.yaml                             | OK                        | 
| ./APIs/mailboxvalidator.com/checker/1.0.0/openapi.yaml        | OK                        | 
| ./APIs/mailboxvalidator.com/disposable/1.0.0/openapi.yaml     | OK                        | 
| ./APIs/mailboxvalidator.com/validation/0.1/openapi.yaml       | OK                        | 
| ./APIs/ably.io/1.1.0/openapi.yaml                             | OK                        | |          
| ./APIs/listennotes.com/2.0/openapi.yaml                       | OK                        | 
| ./APIs/fraudlabspro.com/fraud-detection/1.1/openapi.yaml      | OK                        | 
| ./APIs/fraudlabspro.com/sms-verification/1.0/openapi.yaml     | OK                        | 
| ./APIs/bbci.co.uk/1.0/openapi.yaml                            | OK                        |   
| ./APIs/mashape.com/football-prediction/2/openapi.yaml         | OK                        | |
|./APIs/tomtom.com/maps/1.0.0/openapi.yaml                      | OK                        | 
|./APIs/tomtom.com/search/1.0.0/openapi.yaml                    | gnostic reading error         | 
|./APIs/tomtom.com/routing/1.0.0/openapi.yaml                   | OK                        | |
|./APIs/ip2proxy.com/1.0/openapi.yaml                           | OK                        | 
|./APIs/brex.io/1.0.0/openapi.yaml                              | OK                        | |
|./APIs/adyen.com/AccountService/4/openapi.yaml                 | gnostic reading error         | 
|./APIs/adyen.com/AccountService/3/openapi.yaml                 | gnostic reading error         | 
|./APIs/adyen.com/AccountService/5/openapi.yaml                 | gnostic reading error         | 
|./APIs/adyen.com/PayoutService/30/openapi.yaml                 | gnostic reading error         | 
|./APIs/adyen.com/PaymentService/30/openapi.yaml                | gnostic reading error         | 
|./APIs/adyen.com/PaymentService/46/openapi.yaml                | gnostic reading error         | 
|./APIs/adyen.com/PaymentService/40/openapi.yaml                | gnostic reading error         | 
|./APIs/adyen.com/PaymentService/25/openapi.yaml                | gnostic reading error         | 
|./APIs/adyen.com/FundService/3/openapi.yaml                    | gnostic reading error         | 
|./APIs/adyen.com/FundService/5/openapi.yaml                    | gnostic reading error         | 
|./APIs/adyen.com/CheckoutService/32/openapi.yaml               | gnostic reading error         | 
|./APIs/adyen.com/CheckoutService/31/openapi.yaml               | gnostic reading error         | 
|./APIs/adyen.com/CheckoutService/30/openapi.yaml               | gnostic reading error         | 
|./APIs/adyen.com/CheckoutService/37/openapi.yaml               | gnostic reading error         | 
|./APIs/adyen.com/CheckoutService/46/openapi.yaml               | gnostic reading error         | 
|./APIs/adyen.com/CheckoutService/41/openapi.yaml               | gnostic reading error         | 
|./APIs/adyen.com/CheckoutService/40/openapi.yaml               | gnostic reading error         | 
|./APIs/adyen.com/NotificationConfigurationService/1/openapi.yaml| gnostic reading error         | 
|./APIs/adyen.com/CheckoutUtilityService/1/openapi.yaml         | OK                        | 
|./APIs/adyen.com/RecurringService/18/openapi.yaml              | gnostic reading error          | 
|./APIs/adyen.com/RecurringService/25/openapi.yaml              | gnostic reading error          | 
|./APIs/microsoft.com/graph/1.0.1/openapi.yaml                  | OK                        | |
