package mmh

func ConfigExample() []Server {
	return []Server{
		{
			Name:     "prod11",
			User:     "root",
			Group:    "prod",
			Address:  "10.10.4.11",
			Port:     22,
			Password: "password",
		},
		{
			Name:      "prod12",
			User:      "root",
			Group:     "prod",
			Address:   "10.10.4.12",
			Port:      22,
			PublicKey: "/Users/mritd/.ssh/id_rsa",
		},
	}
}
