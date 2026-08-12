package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const natsReminderSubject = "eregen.reminder.medication"

// ReminderSender publishes medication reminders to NATS.
type ReminderSender struct {
	js  nats.JetStreamContext
	log *zap.Logger
}

// NewReminderSender creates a NATS JetStream publisher for reminders.
func NewReminderSender(nc *nats.Conn, log *zap.Logger) (*ReminderSender, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}
	return &ReminderSender{js: js, log: log}, nil
}

// SendReminder publishes a medication reminder event for an elderly person.
func (s *ReminderSender) SendReminder(ctx context.Context, elderlyID, ruleID, message string) error {
	payload := map[string]interface{}{
		"elderly_id": elderlyID,
		"rule_id":    ruleID,
		"message":    message,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal reminder: %w", err)
	}
	_, err = s.js.Publish(natsReminderSubject, data)
	if err != nil {
		return fmt.Errorf("publish reminder: %w", err)
	}
	s.log.Info("reminder published", zap.String("elderly_id", elderlyID), zap.String("rule_id", ruleID))
	return nil
}

// MedicationReminderService checks for upcoming medication due and sends reminders.
type MedicationReminderService struct {
	store          store.Store
	reminderSender *ReminderSender
	log            *zap.Logger
}

// NewMedicationReminderService creates a service for checking and sending medication reminders.
func NewMedicationReminderService(
	st store.Store,
	reminderSender *ReminderSender,
	log *zap.Logger,
) *MedicationReminderService {
	return &MedicationReminderService{
		store:          st,
		reminderSender: reminderSender,
		log:            log,
	}
}

// CheckAndSendReminders finds active medication rules for today and sends reminders for upcoming schedules.
// This should be called periodically (e.g., every minute by a scheduler).
func (s *MedicationReminderService) CheckAndSendReminders(ctx context.Context) error {
	now := time.Now()
	currentTime := now.Format("15:04")
	currentDayOfWeek := int(now.Weekday())
	if currentDayOfWeek == 0 {
		currentDayOfWeek = 7
	}

	// Get all elderly profiles
	elderlyList, err := s.store.ListElderly(ctx, 1, 1000)
	if err != nil {
		return fmt.Errorf("list elderly: %w", err)
	}

	for _, elderly := range elderlyList {
		// Get active medication rules for this elderly
		rules, err := s.store.ListMedicationRules(ctx, elderly.ID)
		if err != nil {
			s.log.Warn("list rules failed", zap.String("elderly_id", elderly.ID), zap.Error(err))
			continue
		}

		for _, rule := range rules {
			// Check if rule applies today
			if !ruleAppliesToday(rule, currentDayOfWeek) {
				continue
			}

			// Check if this schedule time matches current time
			if currentTime != rule.ScheduleTime {
				continue
			}

			// Check if already taken today
			takenToday, err := s.store.GetTodayMedStatus(ctx, elderly.ID)
			if err != nil {
				s.log.Warn("check taken today failed", zap.String("rule_id", rule.ID), zap.Error(err))
				continue
			}

			// Check if this rule was already taken
			alreadyTaken := false
			for _, record := range takenToday {
				if record.RuleID == rule.ID && record.Taken {
					alreadyTaken = true
					break
				}
			}
			if alreadyTaken {
				continue
			}

			// Send reminder
			message := fmt.Sprintf("该服药了：%s %s", rule.PillType, rule.ScheduleTime)
			if err := s.reminderSender.SendReminder(ctx, elderly.ID, rule.ID, message); err != nil {
				s.log.Error("send reminder failed", zap.String("rule_id", rule.ID), zap.Error(err))
			}
		}
	}

	return nil
}

// ruleAppliesToday checks if a medication rule applies on the given day of week.
func ruleAppliesToday(rule model.MedicationRule, currentDayOfWeek int) bool {
	if !rule.Active {
		return false
	}
	for _, day := range rule.DaysOfWeek {
		if day == currentDayOfWeek {
			return true
		}
	}
	return false
}
