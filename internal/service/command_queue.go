package service

import (
	"fmt"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

type CommandQueueService struct {
	repos *repo.Repositories
}

func NewCommandQueueService(repos *repo.Repositories) *CommandQueueService {
	return &CommandQueueService{repos: repos}
}

func (s *CommandQueueService) QueueCommand(serverID int64, command, payload string) (*model.ServerCommand, error) {
	cmd := &model.ServerCommand{
		ServerID: serverID,
		Command:  command,
		Payload:  payload,
	}
	if err := s.repos.Command.Insert(cmd); err != nil {
		return nil, fmt.Errorf("queue command: %w", err)
	}
	return cmd, nil
}

func (s *CommandQueueService) PendingCommands(serverID int64) ([]model.ServerCommand, error) {
	cmds, err := s.repos.Command.PendingByServerID(serverID)
	if err != nil {
		return nil, fmt.Errorf("pending commands: %w", err)
	}
	return cmds, nil
}

func (s *CommandQueueService) CompleteCommand(id int64, result string) error {
	return s.repos.Command.MarkCompleted(id, result)
}

func (s *CommandQueueService) FailCommand(id int64, result string) error {
	return s.repos.Command.MarkFailed(id, result)
}

func (s *CommandQueueService) TakeAndRun(id int64) error {
	return s.repos.Command.MarkRunning(id)
}

func (s *CommandQueueService) CommandHistory(serverID int64, limit int) ([]model.ServerCommand, error) {
	cmds, err := s.repos.Command.ListByServerID(serverID, limit)
	if err != nil {
		return nil, fmt.Errorf("command history: %w", err)
	}
	return cmds, nil
}
