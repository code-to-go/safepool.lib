# Intro
Weshare is a distributed secure add-only content distribution based on passive storage. It comes both as a C and Go libraries

# Key Concepts

## User
A user is a person that intends to distribute data. A user is identified by a public/private key (ed25519). By extension a user is a software process run (potentially in background) under the identity of a user.

## Local Storage
The local storage is a memory space on a device owned by a user. For performance reasons usually data is kept in a local storage

## Exchange
An exchange is a location where data and changes are stored so to be asynchronously shared across users. 
An exchange is implemented with existing technologies, such as SFTP, S3 and Azure Storage.


## Users
A user is identified by a pair private/public key. A user can be either an admin or a follower.
- admin: he can add/remove users
- follower: he can share and download data but he cannot add/remove users

## Domain
A domain identifies the users who can share specific data. Each user in a domain is identified by its private/public key.
Domains have a hierarchical structure similar to Internet domains (e.g. public.weshare.zone). In the future this hierarchy may be used for shared access


## Lineage
It is the sequence of changes applied to a file since its creation. 

## Beauty Context
If multiple users modify the same file, the lineage has a fork. When a fork is present, a user must choose his favorite version by downloading it. The vote 

## Access
Data located in exchanges is subject to access control. Since exchanges cannot offer active control, the access is passive by encryption. 

_While all clients that access a exchanger can see the content, only entitled clients can decrypt specific content_

In fact each file must be encrypted with a simmetric key (_AES256_)

## Synchronization
This is the core operation when a client receives updates from the network and uploads possible changes. It is defined in multiple phases

### 1. Local discovery
Files for each domain are checked against information on the DB. If no information about the file is on the DB, the hash is calculated so to check for rename cases. 
In case of rename, the record is update with the status _UPDATE_.

If the DB already contains information about the file, the modification time and the content are checked looking for changes; in case of changes, the status is set to _UPDATE_.

### 2. Remote discovery
A client connect to the closest exchange (round-robin latency is used) available. Then files are filtered based on the Snowfallid, ignoring all files that are older according to the logs in the DB.

For each change file:
- the client rebuilds the chain of changes 


# API

```
  char* start() 
  func AddDomain(domain string) error
  func AddExchange(domain string, exchange json) error
  func GetPublic() string

```




# Design
- Layer1: Storage
- Layer2: Encryption
- Layer3: Feeds

A exchanger is defined by the following folders:
- _data_: contains the exchanger content
- _users_: contains a file for each user with his public key and 

## Local 
Each client keeps some information locally. Most data is stored in a SQLite db.

### TABLE Config
contains configuration parameters both at global and domain level

| Field | Type | Constraints | Description |
|------|----|----|-----------|
| domain | VARCHAR(128) |  | Domain the configuration refers to. When the config is global, the value is NULL |
| key | VARCHAR(64) | NOT NULL | Key of the config |
| value | VARCHAR(64) | NOT NULL | Value of the config |


The following config parameters are supported:
- identity.public: public key of the user
- identity.private: private key of the user
- 

### TABLE Log
tracks all change coming from the net

| Field | Type | Constraints | Description |
|------|----|----|-----------|
| domain | VARCHAR(128) |  | Domain |
| name | VARCHAR(128) |  | Full path of the file |
| hash | CHAR(64) | NOT NULL | Hash of the file |
| change | VARCHAR(16) | NOT NULL | Change file on the network |



### TABLE Files
links names on the file system and their hash value

| Field | Type | Constraints | Description |
|------|----|----|-----------|
| name | VARCHAR(8192) |  | Name of the file |
| hash | CHAR(64) | NOT NULL | Hash of the file |
| modtime | TEXT| NOT NULL | Last modification time of the file |
| status | CHAR(16)| NOT NULL | CREATE,UPDATE,SYNC |
| push | BOOL | NOT NULL | True when the file is to be pushed
| merkle | BLOB| NOT NULL | CREATE,UPDATE,SYNC |



### TABLE Merkle
Store the 


### Invite
The invite is the way to access a domain. The invite contains one or more exchange credentials and the administrators of the group (their public key). It is encrypted with the public key of the receiver.  

| Field | Type | Size (bits)| Content |
|------|----|----|-----------|
| version | uint | 16 | version of the file format, 1.0 at the moment|
| admin | byte[32] | 32| public key of the admin that created the invite|
| config | byte[n] | variable| exchanges configuration in json format|

### Users


## Remote layout
The remote storage in a single folder named after the domain and contains the following files.
In the below description:9,223,372,036,854,775,8
- x is a snowflake id
- n is a numeric split id. I

All files have a version id, which is 1.0

```
📦public.weshare
┣📜U.
┃ ┣📜admins
┃ ┣📜users
┃ ┣📜exchanges
┃ ┣📜log
┃ ┗ 📂merkle
┃   ┣📜d4ccaf2627557c756a0762419a4b6695ddef78dd8c9f78dd8c93262755726273.mrk
┃   ┗📜3a42c503953909637f78dd8c99b3b85ddde362415585afc11901bdefe8349102.mrk
┣ 📜README.md
┣ 📂manual
  ┗📜index.md
```


### U.x
A user file defines the users that have current access to the domain

| Field | Type | Size (bits)| Content |
|------|----|----|-----------|
| version | uint | 16 | version of the file format, 1.0 at the moment|
| users | User[] | variable| list of users|

and each user consists of 

| Field | Type | Size (bits)| Content |
|------|----|----|-----------|
| public | []byte | 128 | ed25519 public key|
| flags | uint | 16 | reserved must be 0|
| aes | string | variable | symmetric encryption key used 
| name | string | variable | first name of the user|
| name2 | string | variable | second name of the user (used in case of multiple users with the same name)|


### C.x 
A change file contains an update on a file. It is made of

| Field | Type | Size (bits)| Content |
|------|----|----|-----------|
| version | uint | 16 | version of the file format, 1.0 at the moment|
| headerSize | uint | 32 | size of the header&#x00B9;, i.e. all the fields except|
| names | string[] | variable | list of names in local |
| origin | string | variable | list of names in exchanger |
| xorHash | byte[] | 256 | xor hash of all parts hashes |
| hashes | byte[][] | variable (x256) | hashes for each part of content&#x00B2; |
| message | string | variable | optional markdown message for other users before they receive the change |
| changes | Change[] | variable | changes against the origin file |
| data | byte[] | variable | the actual data |

Each Change is made in fact of
| Field | Type | Size (bits)| Content |
|------|----|----|-----------|
| type | uint | 16 | type of change: create, replace, delete, insert|
| from | uint | 32 | size of the header&#x00B9;, i.e. all the fields except|
| from | uint | 32 | size of the header&#x00B9;, i.e. all the fields except|

&#x00B9; All the fields before data are the file header

&#x00B2; Parts are built with a Hashsplit algorithm



### A.x
Action file. It defines actions each user can request to the other users. This includes:
- Truncate: delete oldest change files. This usually requires merge of oldest files with latest changes

### Sign and encryption

| File | Signed | Encrypted |
|------|--------|-----------|
| Group | &#10004;  | |
| C.x | &#10004; | &#10004; |
| K.x | &#10004; | |
| N.n | &#10004; | &#10004; |
| N.n | &#10004; | |
| A.x | &#10004; | &#10004; |

Signing is implemented with a ed25519 signing where the public/private keys are the identity of each user. 
On the file system, both the signer public key and the signature are added after the content in binary form

| Field | Type | Size (bits)| Content |
|------|----|----|-----------|
| public | uint | 256 | ed|
| hash | uint | 256 | hash value|


Encryption is implemented with AES256. 