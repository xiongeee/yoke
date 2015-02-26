package main

import(
	"bufio"
	// "errors"
	"fmt"
	"os"
	// "strings"
)

//
func things(st *Status) string {
	if st.CRole == "primary" {
		return "secondary"
	}
	return "primary"
}

//
func configureHBAConf() error {

	self, _ := Whoami()
  other, _ := Whois(otherRole(self))

	//
	entry := fmt.Sprintf(`
host    replication     postgres        %s            trust
`, other.Ip)

	file := conf.DataDir+"pg_hba.conf"

	//
	// fi, err := stat(dataRoot"pg_hba.conf")

	//
	f, err := os.Create(file)
	if err != nil {
		log.Error("[pg_config.configureHBAConf] Failed to create '%s'!\n%s\n", file, err)
		return err
	}

	//
	if _, err := f.WriteString(entry); err != nil {
		log.Error("[pg_config.configureHBAConf] Failed to write to '%s'!\n%s\n", file, err)
		return err
	}

	return nil
}

//
func configurePGConf(master bool) error {

	// # 80 GB required on pg_xlog

	//
	entry := `
wal_level = hot_standby
archive_mode = on
archive_command = 'exit 0'
max_wal_senders = 10
wal_keep_segments = 5000
hot_standby = on`

	// master only
	if master {
		entry += `
			synchronous_standby_names = slave`
	}

	file := conf.DataDir+"postgres.conf"

	//
	f, err := os.Create(file)
	if err != nil {
		log.Error("[pg_config.configurePGConf] Failed to create '%s'!\n%s\n", file, err)
		return err
	}

	//
	if _, err := f.WriteString(entry); err != nil {
		log.Error("[pg_config.configurePGConf] Failed to write to '%s'!\n%s\n", file, err)
		return err
	}

	// #wal_level = minimal                    # minimal, archive, or hot_standby
	//                                         # (change requires restart)
	// #archive_mode = off             # allows archiving to be done
	//                                 # (change requires restart)
	// #archive_command = ''           # command to use to archive a logfile segment
	//                                 # placeholders: %p = path of file to archive
	//                                 #               %f = file name only
	//                                 # e.g. 'test ! -f /mnt/server/archivedir/%f && cp %p /mnt/server/archivedir/%f'
	// #max_wal_senders = 0            # max number of walsender processes
	//                                 # (change requires restart)
	// #wal_keep_segments = 0          # in logfile segments, 16MB each; 0 disables
	// #hot_standby = off                      # "on" allows queries during recovery
	//                                         # (change requires restart)

	//
	// opts := make(map[string]string)
	// opts["wal_level"] 								= "hot_standby"
	// opts["archive_mode"] 							= "on"
	// opts["archive_command"] 					= "exit 0"
	// opts["max_wal_senders"] 					= "10"
	// opts["wal_keep_segments"] 				= "5000"
	// opts["hot_standby"] 							= "on"
	// opts["synchronous_standby_names"] = "slave"

	return nil
}

//
func createRecovery() error {

	file := conf.DataDir+"recovery.conf"
	self, _ := Whoami()
  other, _ := Whois(otherRole(self))

	//
	f, err := os.Create(file)
	if err != nil {
		log.Error("[pg_config.createRecovery] Failed to create '%s'!\n%s\n", file, err)
		return err
	}

	//
	entry := fmt.Sprintf(`# -------------------------------------------------------
# PostgreSQL recovery config file generated by Pagoda Box
# -------------------------------------------------------

# When standby_mode is enabled, the PostgreSQL server will work as a standby. It
# tries to connect to the primary according to the connection settings
# primary_conninfo, and receives XLOG records continuously.
standby_mode = on
primary_conninfo = 'host=%s port=%s application_name=slave'

# restore_command specifies the shell command that is executed to copy log files
# back from archival storage. This parameter is *required* for an archive
# recovery, but optional for streaming replication. The given command satisfies
# the requirement without doing anything.
restore_command = 'exit 0'`, other.Ip, other.PGPort)

	//
	if _, err := f.WriteString(entry); err != nil {
		log.Error("[pg_config.createRecovery] Failed to write to '%s'!\n%s\n", file, err)
		return err
	}

	return nil
}

//
func destroyRecovery() {

	file := conf.DataDir+"recovery.conf"

	//
	err := os.Remove(file)
	if err != nil {
		log.Warn("[pg_config.destroyRecovery] No recovery.conf found at '%s'", file)
	}
}

//
func stat(f string) (os.FileInfo, error) {
	fi, err := os.Stat(f)
	if err != nil {
		log.Fatal("[pg_config.readFile]", err)
		return nil, err
	}

	return fi, nil
}

// parseFile will parse a config file, returning a 'opts' map of the resulting
// config options.
func parseFile(file string) (map[string]string, error) {

	// attempt to open file
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	conf := make(map[string]string)
	scanner := bufio.NewScanner(f)
	readLine := 1

	// Read line by line, sending lines to parseLine
	for scanner.Scan() {
		if err := parseLine(scanner.Text(), conf); err != nil {
			log.Error("[pg_config] Error reading line: %v\n", readLine)
			return nil, err
		}

		readLine++
	}

	return conf, nil
}

// parseLine reads each line of the config file, extracting a key/value pair to
// insert into an 'conf' map.
func parseLine(line string, conf map[string]string) error {

	// if the line isn't already in the map add it
	if _, ok := conf[line]; !ok {
		conf[line] = line
	}

	return nil
}