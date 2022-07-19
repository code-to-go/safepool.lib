package access

type Group struct {
	Name   string
	Admins []PublicKey
	Users  []PublicKey
}

func NewGroup(name string, identity Identity) Group {
	return Group{
		Name:   name,
		Admins: []PublicKey{identity.Public},
		Users:  []PublicKey{identity.Public},
	}
}

type GroupEnvelope struct {
	Data   []byte
	Sig    []byte
	Signed PublicKey
}

// func WriteGroup(group Group, identity Identity, s stores.Storer) error {
// 	data, err := yaml.Marshal(group)
// 	if errors.IsErr(err, "cannot marshal group %s: %v", group.Name, err) {

// 	}
// 	sign, err := Sign(identity, data)

// }
