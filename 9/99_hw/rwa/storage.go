package main

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

type Profile struct {
	ID        string `json:"id,omitempty" gorm:"primary_key;autoIncrement"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Bio       string `json:"bio,omitempty"`
	Image     string `json:"image,omitempty"`
	CreatedAt string `json:"createdAt,omitempty" gorm:"column:created_at"`
	UpdatedAt string `json:"updatedAt,omitempty" gorm:"column:updated_at"`
	Token     string `json:"token,omitempty"`
	Following bool
}

type Article struct {
	ID             string   `gorm:"primary_key;autoIncrement"`
	AuthorID       string   `gorm:"column:author_id"`
	Body           string   `json:"body"`
	Description    string   `json:"description"`
	Favorite       bool     `json:"favorited" gorm:"column:favorited"`
	FavoritesCount int      `json:"favoritesCount" gorm:"column:favorites_count"`
	Slug           string   `json:"slug"`
	TagList        []string `json:"tagList" gorm:"-"`
	Title          string   `json:"title"`
	CreatedAt      string   `json:"createdAt,omitempty" gorm:"column:created_at"`
	UpdatedAt      string   `json:"updatedAt,omitempty" gorm:"column:updated_at"`
	Author         Profile  `json:"author" gorm:"foreignKey:AuthorID"`
}

type Tag struct {
	ID   int    `json:"id,omitempty" gorm:"primary_key;autoIncrement"`
	Name string `json:"name"`
}

type ArticleTag struct {
	TagID     int    `json:"tag_id" gorm:"column:tag_id"`
	ArticleID string `json:"article_id" gorm:"column:article_id"`
}

func (sm *ServerManager) GetProfileByStruct(oldProfile *Profile) *Profile {
	var profile Profile
	sm.db.Where(sm.db.Where("username = ? AND password = ?", oldProfile.Username, oldProfile.Password).Or("email = ? AND password = ?", oldProfile.Email, oldProfile.Password)).First(&profile)
	return &profile
}

func (sm *ServerManager) RegisterProfile(profile *Profile) error {
	var count int64

	if err := sm.db.Model(&Profile{}).Where("email = ? OR username = ?", profile.Email, profile.Username).Count(&count).Error; err != nil {
		return fmt.Errorf("error checking profile existence: %v", err)
	}

	if count > 0 {
		return fmt.Errorf("profile with email or username already exists")
	}

	sessID := RandStringRunes(32)
	profile.Token = sessID

	if err := sm.db.Create(&profile).Error; err != nil {
		return fmt.Errorf("error creating profile: %v", err)
	}

	sm.InsertSession(sessID, profile.ID)

	fmt.Printf("[RegisterProfile] %#v\n", profile)

	return nil
}

func (sm *ServerManager) CheckPassword(profile *Profile) (*Profile, error) {
	oldProfile := sm.GetProfileByStruct(profile)
	if oldProfile == nil {
		return nil, fmt.Errorf("Profile with id %s does not exist", profile.ID)
	}

	if oldProfile.Password != profile.Password {
		return nil, fmt.Errorf("password for profile with id %s does not match", profile.ID)
	}
	return oldProfile, nil
}

func (sm *ServerManager) GetUserByToken(token string) (*Profile, error) {
	var profile *Profile

	sm.db.First(&profile, "token = ?", token)
	if profile.ID == "" {
		return nil, fmt.Errorf("Profile with token %s does not exist", token)
	}

	return profile, nil
}

func (sm *ServerManager) InsertSession(sessID, profileID string) {
	intProfileID, err := strconv.Atoi(profileID)
	if err != nil {
		panic(err)
	}
	session := Session{
		ProfileID: uint32(intProfileID),
		SessionID: sessID,
	}
	sm.db.Create(&session)
}

func (sm *ServerManager) DeleteSessionDB(profileID uint32) {
	sm.db.Delete(&Session{ProfileID: profileID})
}

func (sm *ServerManager) GetSessionBySessID(sessID string) (*Session, error) {
	var session *Session
	err := sm.db.First(&session, "session_id = ?", sessID).Error
	if err != nil {
		return nil, fmt.Errorf("Session with id %s does not exist", sessID)
	}
	return session, nil
}

func (sm *ServerManager) UpdateUserDB(newParams *Profile, token string) (*Profile, error) {
	oldUser, err := sm.GetUserByToken(token)
	fmt.Printf("Before update - %#v\n", oldUser)
	if err != nil {
		return nil, err
	}

	oldVal := reflect.ValueOf(oldUser).Elem()
	newVal := reflect.ValueOf(newParams).Elem()

	for i := 0; i < oldVal.NumField(); i++ {
		oldField := oldVal.Field(i)
		newField := newVal.Field(i)

		if !reflect.DeepEqual(newField.Interface(), reflect.Zero(newField.Type()).Interface()) {
			oldField.Set(newField)
		}
	}

	fmt.Printf("After update - %#v\n", oldUser)
	oldUser.UpdatedAt = time.Now().Format(time.RFC3339)
	sm.db.Updates(&oldUser)
	fmt.Printf("[UpdateUser] %#v/n", newParams)

	return oldUser, nil
}

func (sm *ServerManager) LogoutDB(token string) {
	sm.db.Delete(&Profile{}, "token = ?", token)
}

func (sm *ServerManager) GetAllArticles(query url.Values) ([]Article, error) {
	var articles []Article

	if err := sm.db.Preload("Author").Find(&articles).Error; err != nil {
		return nil, fmt.Errorf("Error getting all articles: %v", err)
	}

	for i := range articles {
		var tags []Tag
		var articleTags []ArticleTag
		if err := sm.db.Where("article_id = ?", articles[i].ID).Find(&articleTags).Error; err != nil {
			return nil, fmt.Errorf("Error getting tags for article: %v", err)
		}

		for _, articleTag := range articleTags {
			var tag Tag
			if err := sm.db.First(&tag, articleTag.TagID).Error; err != nil {
				return nil, fmt.Errorf("Error getting tag by id: %v", err)
			}
			tags = append(tags, tag)
		}

		for _, tag := range tags {
			articles[i].TagList = append(articles[i].TagList, tag.Name)
		}

		articles[i].Author.CreatedAt = ""
		articles[i].Author.UpdatedAt = ""
		articles[i].Author.Email = ""
	}

	return articles, nil
}

func (sm *ServerManager) CreateArticleDB(article *Article, authorID string) {
	article.AuthorID = authorID
	article.CreatedAt = time.Now().Format(time.RFC3339)
	article.UpdatedAt = time.Now().Format(time.RFC3339)
	article.Slug = RandStringRunes(32)

	sm.db.Create(&article)

	for _, tag := range article.TagList {
		var existingTag Tag

		var count int64
		if err := sm.db.Where("name = ?", tag).First(&existingTag).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			newTag := Tag{Name: tag}
			sm.db.Create(&newTag)
			existingTag = newTag
		}

		sm.db.Model(&ArticleTag{}).Where("tag_id = ? AND article_id = ?", existingTag.ID, article.ID).Count(&count)
		if count == 0 {
			newArticleTag := ArticleTag{
				TagID:     existingTag.ID,
				ArticleID: article.ID,
			}
			sm.db.Create(&newArticleTag)
		}
	}
}
