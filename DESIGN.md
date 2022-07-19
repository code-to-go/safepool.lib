# Intro
Baobab is a distributed secure add-only content distribution based on passive storage

# Key Concepts

## Stores
A store is a service that allows clients to store and retrieve data. 
A store is implemented with existing technologies, such as SFTP, S3 and Azure Storage.

## Users
A user is identified by a pair private/public key. 

## Group
A group is a set of users. All users in the same group have access.

Groups can have a hierarchical structure similar to Internet domains. For instance the 

## Access
Data located in stores is subject to access control. Since stores cannot offer active control, the access is passive by encryption. 

_While all clients that access a store can see the content, only entitled clients can decrypt specific content_

In fact each file must be encrypted with a simmetric key (_AES256_)



# Design
- Layer1: Storage
- Layer2: Encryption
- Layer3: Feeds

A store is defined by the following folders:
- _data_: contains the store content
- _users_: contains a file for each user with his public key and 


## Data layout
The log is build on the storage in a single folder and contains the following files: log.x, 

### Log.x 
Log files contain the updates coming from different clients. Each file contains up to 1024 record, each one representing an update. 
File has an header with: 
- version - 2 bytes
- 

Each record consists of:
- timestamp - 4 bytes 64bit unix epoch
- file id (hash of the file) - 8 bytes
- user id - 16 bytes

x is sequential number


### D.hash.dat
Data file. Hash is the Blake hash of the complete file. When an update is loaded, only changes are loaded. 
Later and anyway before 7 days, the file is rebuilt with all the content

Header consists of:
- version - 2 bytes
- origin - 8 bytes hash (all 0 in case of )
- name - 2 bytes size + data
- user
- options - 2 bytes 
- length - 4 bytes
- blocks each one made of:
  - original offset
  - original size
  - pos
  - new size
- data
- signature


### Index.x
Contains the names and their related hash ids in a BTree or HTree structure


