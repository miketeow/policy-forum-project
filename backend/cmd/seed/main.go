package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype" // Often required for nullable UUIDs in sqlc
	"github.com/jackc/pgx/v5/pgxpool"

	"policy-forum-backend/internal/auth"
	"policy-forum-backend/internal/store"
)

func main() {
	log.Printf("🌱 Starting massive database seeding process...")

	dsn := "postgres://admin:password123@localhost:5432/policy_forum?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("❌ Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	queries := store.New(pool)
	ctx := context.Background()

	log.Printf("🧹 Sweeping old data TRUNCATE CASCADE...")
	_, err = pool.Exec(ctx, "TRUNCATE TABLE users, posts, comments, post_votes, comment_votes CASCADE")
	if err != nil {
		log.Fatalf("❌ Failed to truncate tables: %v", err)
	}

	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		log.Fatalf("❌ Failed to hash password: %v", err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// ==========================================
	// 1. SEED USERS (10 Users)
	// ==========================================
	var userIDs []uuid.UUID
	users := []struct{ Name, Email string }{
		{"Ahmad Fazli", "ahmad.fazli@example.com"},
		{"Siti Nurhaliza", "siti.n@example.com"},
		{"Wei Jie Tan", "weijie.tan@example.com"},
		{"Mei Ling", "mei.ling@example.com"},
		{"Priya Sharma", "priya.s@example.com"},
		{"Muthu Kumar", "muthu.k@example.com"},
		{"Arif Rahman", "arif.r@example.com"},
		{"Chong Wei", "chong.wei@example.com"},
		{"Nurul Ain", "nurul.ain@example.com"},
		{"Kevin Raj", "kevin.raj@example.com"},
	}

	for _, u := range users {
		userID := uuid.New()
		now := time.Now().UTC()
		_, err := queries.CreateUser(ctx, store.CreateUserParams{
			ID:             userID,
			Name:           u.Name,
			Email:          u.Email,
			HashedPassword: hashedPassword,
			KycStatus:      "VERIFIED",
			CreatedAt:      now,
			UpdatedAt:      now,
		})
		if err != nil {
			log.Fatalf("Failed to insert user: %v", err)
		}
		userIDs = append(userIDs, userID)
	}
	log.Printf("✅ Inserted 10 users")

	// ==========================================
	// 2. SEED POSTS (20 Posts, 5 Categories)
	// ==========================================
	var postIDs []uuid.UUID
	categories := []string{"INFRASTRUCTURE", "ECONOMY", "HEALTHCARE", "EDUCATION", "ENVIRONMENT"}

	for i := 0; i < 20; i++ {
		postID := uuid.New()
		authorID := userIDs[rng.Intn(len(userIDs))]
		category := categories[i%len(categories)] // Rotates evenly through categories

		_, err := queries.CreatePost(ctx, store.CreatePostParams{
			ID:        postID,
			UserID:    authorID,
			Title:     fmt.Sprintf("Discussion Thread #%d: %s Impact", i+1, category),
			Content:   "This is a seeded discussion post designed to test our new filtering, search, and categorization logic. We need to evaluate the long-term impact of these policies on our local communities.",
			Category:  store.PostCategory(category),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			// Note: Depending on your sqlc setup, you might need to supply Score: 0 here
		})
		if err != nil {
			log.Fatalf("Failed to insert post: %v", err)
		}
		postIDs = append(postIDs, postID)
	}
	log.Printf("✅ Inserted 20 posts")

	// ==========================================
	// 3. SEED ROOT COMMENTS (10 per Post = 200)
	// ==========================================
	var rootCommentIDs []uuid.UUID
	genericComments := []string{
		"I completely agree with the points raised here.",
		"This requires more transparency from the local council.",
		"Have they considered the budget constraints for this?",
		"Great initiative, but the execution timeline is unrealistic.",
		"This will negatively impact lower-income areas if not planned properly.",
	}

	for _, postID := range postIDs {
		for i := 0; i < 10; i++ {
			commentID := uuid.New()
			authorID := userIDs[rng.Intn(len(userIDs))]
			content := genericComments[rng.Intn(len(genericComments))]

			_, err := queries.CreateComments(ctx, store.CreateCommentsParams{
				ID:        commentID,
				PostID:    postID,
				UserID:    authorID,
				Content:   content,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				// ParentID is left empty/null for root comments
			})
			if err != nil {
				log.Fatalf("Failed to insert root comment: %v", err)
			}
			rootCommentIDs = append(rootCommentIDs, commentID)
		}
	}
	log.Printf("✅ Inserted 200 root comments")

	// ==========================================
	// 4. SEED NESTED COMMENTS (10 replies to 3 specific roots)
	// ==========================================
	// We pick the first 3 root comments to be the "hot" threads
	for i := 0; i < 3; i++ {
		targetParentID := rootCommentIDs[i]
		// We need to fetch the post_id of the parent comment to maintain data integrity
		// For simplicity in the seed script, we know rootCommentIDs[0...9] belong to postIDs[0]
		targetPostID := postIDs[0]

		for j := 0; j < 10; j++ {
			authorID := userIDs[rng.Intn(len(userIDs))]

			// Constructing the nullable UUID for sqlc
			parentUUID := pgtype.UUID{
				Bytes: targetParentID,
				Valid: true,
			}

			_, err := queries.CreateComments(ctx, store.CreateCommentsParams{
				ID:        uuid.New(),
				PostID:    targetPostID,
				UserID:    authorID,
				Content:   fmt.Sprintf("Reply #%d to the main argument. I see your point.", j+1),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				ParentID:  parentUUID,
			})
			if err != nil {
				log.Fatalf("Failed to insert nested comment: %v", err)
			}
		}
	}
	log.Printf("✅ Inserted 30 nested comments")

	// ==========================================
	// 5. SEED VOTES (Using the Slice Window technique)
	// ==========================================
	voteScores := []int{10, 10, 8, 8, 6, 6}

	// Post Votes
	for i, score := range voteScores {
		targetPost := postIDs[i]

		// Take exactly 'score' amount of distinct users
		voters := userIDs[:score]
		for _, voterID := range voters {
			// 1. Insert into post_votes table
			// Replace with your actual store.CreatePostVoteParams
			_, err := pool.Exec(ctx, `
				INSERT INTO post_votes (post_id, user_id, vote, created_at)
				VALUES ($1, $2, 1, $3)`,
				targetPost, voterID, time.Now().UTC())

			if err != nil {
				log.Fatalf("Failed to insert post vote: %v", err)
			}
		}

		// 2. Update the denormalized score on the posts table
		_, err = pool.Exec(ctx, "UPDATE posts SET score = $1 WHERE id = $2", score, targetPost)
		if err != nil {
			log.Fatalf("Failed to update post score: %v", err)
		}
	}
	log.Printf("✅ Seeded Post Votes")

	// Comment Votes
	for i, score := range voteScores {
		targetComment := rootCommentIDs[i]

		voters := userIDs[:score]
		for _, voterID := range voters {
			// 1. Insert into comment_votes table
			_, err := pool.Exec(ctx, `
				INSERT INTO comment_votes (comment_id, user_id, vote, created_at)
				VALUES ($1, $2, 1, $3)`,
				targetComment, voterID, time.Now().UTC())

			if err != nil {
				log.Fatalf("Failed to insert comment vote: %v", err)
			}
		}

		// 2. Update the denormalized score on the comments table
		_, err = pool.Exec(ctx, "UPDATE comments SET score = $1 WHERE id = $2", score, targetComment)
		if err != nil {
			log.Fatalf("Failed to update comment score: %v", err)
		}
	}
	log.Printf("✅ Seeded Comment Votes")

	log.Printf("🎉 Database perfectly seeded with Posts, Comments, and Votes!")
}
