## OpenAPI v3
Command to find OpenAPI v3 descriptions inside [APIs-guru/openapi-directory](https://github.com/APIs-guru/openapi-directory):
    
     find . -name "openapi.yaml" -exec gnostic --grpc-out=. {} \;

|APIs                                               | Success       | gnostic fails  | gnostic-grpc fails |
| --------------------------------------------------|:-------------:|:--------------:|:------------------:|
| All APIs inside: APIs-guru/openapi-directory/APIs/| 39            | 25             |0|

### gnostic fails:
Gnostic throws an error while converting the .yaml to Proto Bufs. We see that as long as gnostic converts the OpenAPI 
description, the plugin gnostic-grpc works.

./APIs/bunq.com/1.0/openapi.yaml  
./APIs/linode.com/4.0.15/openapi.yaml  
./APIs/cpy.re/peertube/1.3.1/openapi.yaml  
./APIs/twitter.com/labs/1.1/openapi.yaml  
./APIs/tomtom.com/search/1.0.0/openapi.yaml  
./APIs/adyen.com/AccountService/4/openapi.yaml   
./APIs/adyen.com/AccountService/3/openapi.yaml   
./APIs/adyen.com/AccountService/5/openapi.yaml   
./APIs/adyen.com/PayoutService/30/openapi.yaml   
./APIs/adyen.com/PaymentService/30/openapi.yaml   
./APIs/adyen.com/PaymentService/46/openapi.yaml   
./APIs/adyen.com/PaymentService/40/openapi.yaml   
./APIs/adyen.com/PaymentService/25/openapi.yaml   
./APIs/adyen.com/FundService/3/openapi.yaml   
./APIs/adyen.com/FundService/5/openapi.yaml   
./APIs/adyen.com/CheckoutService/32/openapi.yaml   
./APIs/adyen.com/CheckoutService/31/openapi.yaml   
./APIs/adyen.com/CheckoutService/30/openapi.yaml   
./APIs/adyen.com/CheckoutService/37/openapi.yaml   
./APIs/adyen.com/CheckoutService/46/openapi.yaml   
./APIs/adyen.com/CheckoutService/41/openapi.yaml   
./APIs/adyen.com/CheckoutService/40/openapi.yaml   
./APIs/adyen.com/NotificationConfigurationService/1/openapi.yaml   
./APIs/adyen.com/RecurringService/18/openapi.yaml
./APIs/adyen.com/RecurringService/25/openapi.yaml

### gnostic-grpc fails:
None  

## OpenAPI v2
Command to find OpenAPI v2 descriptions inside [APIs-guru/openapi-directory](https://github.com/APIs-guru/openapi-directory):
    
     find . -name "swagger.yaml" -exec gnostic --grpc-out=. {} \;
     
|APIs                                               | Success       | gnostic fails  | gnostic-grpc fails |
| --------------------------------------------------|:-------------:|:--------------:|:------------------:|
| All APIs inside: APIs-guru/openapi-directory/APIs/| 2461          | 3              |208|
| APIs-guru/openapi-directory/APIs/azure.com/       | 1452          | 0              |207|
| Every folder except azure.com                     | 1009          | 3              |1|

### gnostic fails:
./APIs/slack.com/1.2.0/swagger.yaml  
./APIs/bungie.net/2.0.0/swagger.yaml  
./APIs/rebilly.com/2.1/swagger.yaml  

### gnostic-grpc fails:

####azure.com   
Most of the 207 azure APIs seem to fail, because of following error:

**Example API:** ./APIs/azure.com/network-networkSecurityGroup/2017-06-01/swagger.yaml  
**Example line inside swagger.yaml:** 1754  
**Planning to fix:** No  
**Error description:**
        
    A reference to another file that is not at the corresponding location. 
    
    Example: having a ref like: $ref: './virtualNetwork.json#/definitions/Subnet' (line 1754)
    results in an error thrown by the plugin, because the mentioned file is not there (virtualNetwork.json)
    
#### Every folder except azure.com
Only one API is failing:  

**Example API:** ./APIs/twinehealth.com/v7.78.1/swagger.yaml  
**Example line inside swagger.yaml:** 3088  
**Planning to fix:** No  
**Error description:**

    A reference to the properties of an object.
        
    Example: Having a ref like: $ref: '#/definitions/CalendarEventResource/properties/type' (line 3088)
    results in an error thrown by the plugin, because it is not possible to reference properties of definitions.
