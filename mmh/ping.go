package mmh

func (s *ServerConfig) Ping() error {
	client, err := s.sshClient(true)
	if err != nil {
		return err
	}
	client.Dial()
}
