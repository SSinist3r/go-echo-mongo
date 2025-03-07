package service

import (
	"context"
	"fmt"
	"log"

	"go-echo-mongo/internal/model"
	"go-echo-mongo/internal/repository"
	"go-echo-mongo/internal/repository/redisrepo"
	"go-echo-mongo/pkg/secutil"
	"go-echo-mongo/pkg/strutil"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserService defines the interface for user-related business logic
type UserService interface {
	BaseService[*model.User]
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByApiKey(ctx context.Context, apiKey string) (*model.User, error)
	ValidateCredentials(ctx context.Context, email, password string) (*model.User, error)

	// Role management
	AddRoles(ctx context.Context, id string, roles []string) error
	RemoveRoles(ctx context.Context, id string, roles []string) error
	GetUsersByRole(ctx context.Context, role string) ([]*model.User, error)

	// Batch operations
	CreateUsers(ctx context.Context, users []*model.User) error
	FindUsersByFilter(ctx context.Context, filter map[string]interface{}, limit, skip int64) ([]*model.User, error)
	UpdateUsersByFilter(ctx context.Context, filter interface{}, updates interface{}) (int64, error)
	DeleteUsersByIDs(ctx context.Context, ids []string) (int64, error)
}

type userService struct {
	BaseService[*model.User]
	repo  repository.UserRepository
	redis redisrepo.Repository
}

// NewUserService creates a new UserService instance
func NewUserService(repo repository.UserRepository, redis redisrepo.Repository) UserService {
	if repo == nil {
		log.Fatal(ErrNilRepository)
	}
	return &userService{
		BaseService: newBaseService(repo),
		repo:        repo,
		redis:       redis,
	}
}

// Create overrides base Create to add email check and password hashing
func (s *userService) Create(ctx context.Context, user *model.User) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	// Check for existing email
	if existingUser, _ := s.GetByEmail(ctx, user.Email); existingUser != nil {
		return ErrEmailExists
	}

	// Hash password
	hashedPassword, err := secutil.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	// Generate API key
	apiKey, err := strutil.GenerateRandom(32, false, true, true, false)
	if err != nil {
		return err
	}
	user.ApiKey = apiKey

	// Ensure user has at least the basic user role if no roles are specified
	if len(user.Roles) == 0 {
		user.Roles = []string{model.RoleUser}
	}

	return s.BaseService.Create(ctx, user)
}

// Update overrides base Update to handle email uniqueness and password hashing
func (s *userService) Update(ctx context.Context, id string, updates *model.User) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	existingUser, err := s.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// Check email uniqueness if it's being updated
	if updates.Email != "" && updates.Email != existingUser.Email {
		if emailUser, _ := s.GetByEmail(ctx, updates.Email); emailUser != nil {
			return ErrEmailExists
		}
	}

	// Hash new password if provided
	if updates.Password != "" {
		hashedPassword, err := secutil.HashPassword(updates.Password)
		if err != nil {
			return err
		}
		updates.Password = hashedPassword
	}

	return s.BaseService.Update(ctx, id, updates)
}

// GetByEmail retrieves a user by their email
func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	return s.repo.FindByEmail(ctx, email)
}

// GetByApiKey retrieves a user by their API key
func (s *userService) GetByApiKey(ctx context.Context, apiKey string) (*model.User, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	return s.repo.FindByApiKey(ctx, apiKey)
}

// ValidateCredentials validates user credentials and returns the user if valid
func (s *userService) ValidateCredentials(ctx context.Context, email, password string) (*model.User, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	user, err := s.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := secutil.VerifyPassword(user.Password, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// CreateUsers creates multiple users with email uniqueness check and password hashing
func (s *userService) CreateUsers(ctx context.Context, users []*model.User) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if len(users) == 0 {
		return ErrEmptyBatch
	}

	// Check for duplicate emails within the batch
	emails := make(map[string]bool)
	for _, user := range users {
		if emails[user.Email] {
			return ErrEmailExists
		}
		emails[user.Email] = true

		// Check if email already exists in database
		if existingUser, _ := s.GetByEmail(ctx, user.Email); existingUser != nil {
			return ErrEmailExists
		}

		// Hash password
		hashedPassword, err := secutil.HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword

		// Generate API key
		apiKey, err := strutil.GenerateRandom(32, false, true, true, false)
		if err != nil {
			return err
		}
		user.ApiKey = apiKey

		// Ensure user has at least the basic user role if no roles are specified
		if len(user.Roles) == 0 {
			user.Roles = []string{model.RoleUser}
		}
	}

	return s.BaseService.CreateMany(ctx, users)
}

// FindUsersByFilter finds users by filter criteria
func (s *userService) FindUsersByFilter(ctx context.Context, filter map[string]interface{}, limit, skip int64) ([]*model.User, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	// Convert map to BSON filter
	bsonFilter := bson.M{}
	for k, v := range filter {
		if v != "" {
			bsonFilter[k] = v
		}
	}

	// Set options
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}

	return s.BaseService.FindMany(ctx, bsonFilter, opts)
}

// UpdateUsersByFilter updates users based on filter and updates criteria
// It supports two modes:
// 1. When filter is a map[string]interface{} and updates is map[string]interface{}, it applies the same updates to all matched users
// 2. When filter is map[string]map[string]interface{}, it treats the outer map key as user ID and applies specific updates to each user
// Example 1: Per-User Updates (Current Handler Implementation)
// Processing individual updates for each user
//
//	userUpdates := map[string]map[string]interface{}{
//	    "67c030f9cab775964b91c0f5": {
//	        "name": "Updated User 1",
//	        "email": "updated1@example.com",
//	    },
//	    "67c030f9cab775964b91c0f6": {
//	        "name": "Updated User 2",
//	        "email": "updated2@example.com",
//	    },
//	}
//
// count, err := service.UpdateUsersByFilter(ctx, userUpdates, nil)
//
// Example 2: General Filter Updates
// Update all users with a specific email domain
//
//	filter := map[string]interface{}{
//	    "email": bson.M{"$regex": "@oldcompany.com$"},
//	}
//
//	updates := map[string]interface{}{
//	    "company": "New Company Name",
//	    "updated_at": time.Now(),
//	}
//
// count, err := service.UpdateUsersByFilter(ctx, filter, updates)
//
// Example 3: Update Users with Specific Role
// Update all admin users
//
//	filter := map[string]interface{}{
//	    "roles": "admin",
//	}
//
//	updates := map[string]interface{}{
//	    "access_level": 10,
//	    "updated_at": time.Now(),
//	}
//
// count, err := service.UpdateUsersByFilter(ctx, filter, updates)
func (s *userService) UpdateUsersByFilter(ctx context.Context, filter interface{}, updates interface{}) (int64, error) {
	if err := validateContext(ctx); err != nil {
		return 0, err
	}

	var writeModels []mongo.WriteModel

	// Handle different update patterns based on input types
	switch filterType := filter.(type) {
	// Case 1: User ID to updates map (per-user updates)
	case map[string]map[string]interface{}:
		userUpdates := filterType
		if len(userUpdates) == 0 {
			return 0, nil
		}

		// Process each user update
		for id, userUpdates := range userUpdates {
			// Convert string ID to ObjectID
			objID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				return 0, fmt.Errorf("invalid ID format: %s", id)
			}

			// Create an update model for this user
			updateModel := mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": objID}).
				SetUpdate(bson.M{"$set": userUpdates})

			writeModels = append(writeModels, updateModel)
		}

	// Case 2: General filter with common updates for all matched users
	case map[string]interface{}:
		generalFilter := filterType
		generalUpdates, ok := updates.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("updates must be map[string]interface{} when filter is map[string]interface{}")
		}

		// Create BSON filter
		bsonFilter := bson.M{}
		for k, v := range generalFilter {
			// Handle special ID case
			if k == "_id" {
				switch idValue := v.(type) {
				case string:
					// Single ID as string
					objID, err := primitive.ObjectIDFromHex(idValue)
					if err != nil {
						return 0, fmt.Errorf("invalid ID format: %s", idValue)
					}
					bsonFilter["_id"] = objID
				case []string:
					// Array of ID strings
					objectIDs := make([]primitive.ObjectID, 0, len(idValue))
					for _, id := range idValue {
						objID, err := primitive.ObjectIDFromHex(id)
						if err != nil {
							return 0, fmt.Errorf("invalid ID format: %s", id)
						}
						objectIDs = append(objectIDs, objID)
					}
					bsonFilter["_id"] = bson.M{"$in": objectIDs}
				case bson.M:
					// Already in BSON format
					bsonFilter["_id"] = idValue
				case primitive.ObjectID:
					// Already an ObjectID
					bsonFilter["_id"] = idValue
				default:
					return 0, fmt.Errorf("unsupported _id filter type: %T", v)
				}
			} else if v != "" {
				// For non-ID fields, just add to filter if not empty
				bsonFilter[k] = v
			}
		}

		// Create update model
		updateModel := mongo.NewUpdateManyModel().
			SetFilter(bsonFilter).
			SetUpdate(bson.M{"$set": generalUpdates})

		writeModels = append(writeModels, updateModel)

	default:
		return 0, fmt.Errorf("unsupported filter type: %T", filter)
	}

	// If no write models created, return early
	if len(writeModels) == 0 {
		return 0, nil
	}

	// Use base service to execute the bulk update
	return s.BaseService.UpdateMany(ctx, nil, writeModels)
}

// DeleteUsersByIDs deletes multiple users by their IDs
func (s *userService) DeleteUsersByIDs(ctx context.Context, ids []string) (int64, error) {
	if err := validateContext(ctx); err != nil {
		return 0, err
	}

	if len(ids) == 0 {
		return 0, ErrEmptyBatch
	}

	// Convert string IDs to ObjectIDs
	objectIDs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		objectID, err := model.StringToObjectID(id)
		if err != nil {
			continue // Skip invalid IDs
		}
		objectIDs = append(objectIDs, objectID)
	}

	if len(objectIDs) == 0 {
		return 0, ErrEmptyBatch
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}

	return s.BaseService.DeleteMany(ctx, filter)
}

// AddRoles adds roles to a user
func (s *userService) AddRoles(ctx context.Context, id string, roles []string) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if len(roles) == 0 {
		return nil
	}

	user, err := s.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// Add new roles (avoiding duplicates)
	roleMap := make(map[string]bool)
	for _, role := range user.Roles {
		roleMap[role] = true
	}

	hasNewRoles := false
	for _, role := range roles {
		if !roleMap[role] {
			user.Roles = append(user.Roles, role)
			roleMap[role] = true
			hasNewRoles = true
		}
	}

	// Only update if there are new roles
	if hasNewRoles {
		return s.BaseService.Update(ctx, id, user)
	}

	return nil
}

// RemoveRoles removes roles from a user
func (s *userService) RemoveRoles(ctx context.Context, id string, roles []string) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if len(roles) == 0 {
		return nil
	}

	user, err := s.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// Create a map of roles to remove for quick lookup
	removeMap := make(map[string]bool)
	for _, role := range roles {
		removeMap[role] = true
	}

	// Filter out roles to remove
	newRoles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		if !removeMap[role] {
			newRoles = append(newRoles, role)
		}
	}

	// Only update if roles were actually removed
	if len(newRoles) != len(user.Roles) {
		user.Roles = newRoles
		return s.BaseService.Update(ctx, id, user)
	}

	return nil
}

// GetUsersByRole retrieves all users with a specific role
func (s *userService) GetUsersByRole(ctx context.Context, role string) ([]*model.User, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	filter := bson.M{"roles": bson.M{"$in": []string{role}}}
	return s.BaseService.FindMany(ctx, filter, nil)
}
