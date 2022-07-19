# Introduction
Access control is based on passive mechanisms:
- encryption for data protection
- signature for data validation

# Interfaces

    type User {
        PublicKey string
        Nick string
    }

    type Group {
        AdminKeys []string
        MembersKeys []string
        KeyId string
    }

    type Key {
        EncryptedKeys map[string]string        
    }

    type Entry {
        Name [512]byte
        Hashes [16][16]byte
        NameLen int16
        Created time
        Child int
    }

    type EntryBlock {
        Id int
        Entries [512]Entry
    }

    interface Access {
        CreateGroup(name string) (publicKey, private)
        AddToGroup(name string, user string)
        RemoveFromGroup(name string, user string)

        AddDocument(group, name string)
        RemoveDocument(group, name string)
    }