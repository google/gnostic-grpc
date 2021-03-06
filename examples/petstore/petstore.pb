
3.0.0�
Swagger Petstore�This is a sample Petstore server.  You can find
out more about Swagger at
[http://swagger.io](http://swagger.io) or on
[irc.freenode.net, #swagger](http://swagger.io/irc/).
http://swagger.io/terms/"apiteam@swagger.io*=

Apache 2.0/http://www.apache.org/licenses/LICENSE-2.0.html21.0.0`
Ahttps://virtserver.swaggerhub.com/LorenzHW/originalPetstore/1.0.0SwaggerHub API Auto Mocking"�-
�
/pet�*�
petUpdate an existing pet*	updatePet:" 
#/components/requestBodies/PetB[
400

Invalid ID supplied
404

Pet not found
405

Validation exceptionZ 2k
petAdd a new pet to the store*addPet:" 
#/components/requestBodies/PetB
405

Invalid inputZ 
�
/pet/findByStatus�"�
petFinds Pets by statusCMultiple status values can be provided with comma separated strings*findPetsByStatus2�
�
statusquery3Status values that need to be considered for filter @RN
L�array�A
?
=�
available
�
pending
�sold
�string�	availableB��
200�
�
successful operation�
A
application/json-
+
)�array�

#/components/schemas/Pet
@
application/xml-
+
)�array�

#/components/schemas/Pet
400

Invalid status valueZ 
�
/pet/findByTags�"�
petFinds Pets by tags_Muliple tags can be provided with comma separated strings. Use\ \ tag1, tag2, tag3 for testing.*findPetsByTags2B
@
tagsqueryTags to filter by @R
�array�

	�stringB��
200�
�
successful operation�
A
application/json-
+
)�array�

#/components/schemas/Pet
@
application/xml-
+
)�array�

#/components/schemas/Pet
400

Invalid tag valuePZ 
�
/pet/{petId}�"�
petFind pet by IDReturns a single pet*
getPetById2<
:
petIdpathID of pet to return R
�integer�int64B��
200�

successful operationg
2
application/json

#/components/schemas/Pet
1
application/xml

#/components/schemas/Pet
400

Invalid ID supplied
404

Pet not foundZ 2�
pet)Updates a pet in the store with form data*updatePetWithForm2K
I
petIdpath"ID of pet that needs to be updated R
�integer�int64:�
��
�
!application/x-www-form-urlencodedr
p
n�object�b
-
name%
#�string�Updated name of the pet
1
status'
%�string�Updated status of the petB
405

Invalid inputZ :�
petDeletes a pet*	deletePet2 

api_keyheaderR
	�string29
7
petIdpathPet id to delete R
�integer�int64B:
400

Invalid ID supplied
404

Pet not foundZ 
�
/pet/{petId}/uploadImage�2�
petuploads an image*
uploadFile2<
:
petIdpathID of pet to update R
�integer�int64:8
64
2
application/octet-stream

�string�binaryB_]
200V
T
successful operation<
:
application/json&
$"
 #/components/schemas/ApiResponseZ 
�
/store/inventory�"�
store!Returns pet inventories by status+Returns a map of status codes to quantities*getInventoryB_]
200V
T
successful operation<
:
application/json&
$
"�object�

�integer�int32Z 
�
/store/order�2�
storePlace an order for a pet*
placeOrder:a
_
#order placed for purchasing the pet6
4
application/json 

#/components/schemas/OrderB��
200�
�
successful operationk
4
application/json 

#/components/schemas/Order
3
application/xml 

#/components/schemas/Order
400

Invalid Order
�
/store/order/{orderId}�"�
storeFind purchase order by IDgFor valid response try integer IDs with value >= 1 and <= 10.\ \ Other values will generated exceptions*getOrderById2_
]
orderIdpath"ID of pet that needs to be fetched R&
$Y      $@i      �?�integer�int64B��
200�
�
successful operationk
4
application/json 

#/components/schemas/Order
3
application/xml 

#/components/schemas/Order
400

Invalid ID supplied
404

Order not found:�
storeDelete purchase order by IDzFor valid response try integer IDs with positive integer value.\ \ Negative or non-integer values will generate API errors*deleteOrder2\
Z
orderIdpath(ID of the order that needs to be deleted R
i      �?�integer�int64B<
400

Invalid ID supplied
404

Order not found
�
/user�2�
userCreate user,This can only be done by the logged in user.*
createUser:P
N
Created user object5
3
application/json

#/components/schemas/UserB


successful operation
�
/user/createWithArray�2�
user,Creates list of users with given input array*createUsersWithArrayInput:(&
$#/components/requestBodies/UserArrayB


successful operation
�
/user/createWithList�2�
user,Creates list of users with given input array*createUsersWithListInput:(&
$#/components/requestBodies/UserArrayB


successful operation
�
/user/login�"�
userLogs user into the system*	loginUser2;
9
usernamequeryThe user name for login R
	�string2H
F
passwordquery$The password for login in clear text R
	�stringB��
200�
�
successful operation�
L
X-Rate-Limit<
:
"calls per hour allowed by the userB
�integer�int32
N
X-Expires-After;
9
date in UTC when token expiresB
�string�	date-timeE
!
application/json

	�string
 
application/xml

	�string-
400&
$
"Invalid username/password supplied
i
/user/logoutY"W
user'Logs out current logged in user session*
logoutUserB


successful operation
�
/user/{username}�"�
userGet user by user name*getUserByName2\
Z
usernamepath9The name that needs to be fetched. Use user1 for testing. R
	�stringB��
200�
�
successful operationi
3
application/json

#/components/schemas/User
2
application/xml

#/components/schemas/User$
400

Invalid username supplied
404

User not found*�
userUpdated user,This can only be done by the logged in user.*
updateUser2?
=
usernamepathname that need to be updated R
	�string:P
N
Updated user object5
3
application/json

#/components/schemas/UserB= 
400

Invalid user supplied
404

User not found:�
userDelete user,This can only be done by the logged in user.*
deleteUser2D
B
usernamepath!The name that needs to be deleted R
	�stringBA$
400

Invalid username supplied
404

User not found*�
�	
�
Order�
�*
Order�object��

id
�integer�int64

petId
�integer�int64
 
quantity
�integer�int32
#
shipDate
�string�	date-time
M
statusC
A�	placed
�	approved
�
delivered
�string�Order Status

complete
�boolean� 
W
CategoryK
I*

Category�object�1

id
�integer�int64

name
	�string
�
User�
�*
User�object��

id
�integer�int64

username
	�string

	firstName
	�string

lastName
	�string

email
	�string

password
	�string

phone
	�string
0

userStatus"
 �integer�User Status�int32
M
TagF
D*
Tag�object�1

id
�integer�int64

name
	�string
�
Pet�
�*
Pet�name�	photoUrls�object��

id
�integer�int64
-
category!
#/components/schemas/Category

name
:	doggie
�string
5
	photoUrls(
&*
photoUrl(�array�

	�string
<
tags4
2*
tag(�array�

#/components/schemas/Tag
U
statusK
I�
available
�
pending
�sold
�string�pet status in the store
h
ApiResponseY
W�object�K

code
�integer�int32

type
	�string

message
	�string*�
�
Pet�
�
.Pet object that needs to be added to the storeg
2
application/json

#/components/schemas/Pet
1
application/xml

#/components/schemas/Pet
l
	UserArray_
]
List of user objectD
B
application/json.
,
*�array�

#/components/schemas/User:�
�
petstore_auth�

oauth2:u
s
'http://petstore.swagger.io/oauth/dialog"H
)

write:petsmodify pets in your account

	read:petsread your pets
&
api_key

apiKeyapi_key"header:E
petEverything about your Pets"
Find out morehttp://swagger.io:"
storeAccess to Petstore orders:Q
userOperations about user2
Find out more about our storehttp://swagger.ioB0
Find out more about Swaggerhttp://swagger.io