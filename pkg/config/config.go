package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type RingLeaderConfig struct {
	TaskQueueSize uint32            `json:"task_queue_size" toml:"task_queue_size"`
	Connections   ConnectionsConfig `json:"connections" toml:"connections"`
}

type ConnectionsConfig struct {
	MaxRetries         int `json:"max_retries" toml:"max_retries"`
	TimeBetweenRetries int `json:"time_between_retries" toml:"time_between_retries"`
}

type WorkerConfig struct {
	HeartbeatInterval int               `json:"heartbeat_interval" toml:"heartbeat_interval"`
	BackoffMax        int               `json:"backoff_max" toml:"backoff_max"`
	Connections       ConnectionsConfig `json:"connections" toml:"connections"`
}

type TaskSchedulerConfig struct {
	Title            string           `json:"title" toml:"title"`
	RingLeaderConfig RingLeaderConfig `json:"ring_leader" toml:"ring-leader"`
	WorkerConfig     WorkerConfig     `json:"worker" toml:"worker"`
}

var (
	defaultConfig = &TaskSchedulerConfig{
		Title: "task-scheduler",
		RingLeaderConfig: RingLeaderConfig{
			TaskQueueSize: 1024,
			Connections: ConnectionsConfig{
				MaxRetries:         10,
				TimeBetweenRetries: 5, // NOTE: this is in seconds
			},
		},
		WorkerConfig: WorkerConfig{
			HeartbeatInterval: 2,
			BackoffMax:        2,
			Connections: ConnectionsConfig{
				MaxRetries:         10,
				TimeBetweenRetries: 5,
			},
		},
	}
)

func ParseConfig(fileName string) (*TaskSchedulerConfig, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return defaultConfig, err
	}

	var config TaskSchedulerConfig

	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
