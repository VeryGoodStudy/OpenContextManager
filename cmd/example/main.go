// Command example demonstrates the usage of the OpenContextManager library
// with both memory and file storage backends, and context management strategies.
package main

import (
	"fmt"
	"log"

	ocm "github.com/VeryGoodStudy/OpenContextManager"
	"github.com/VeryGoodStudy/OpenContextManager/store/memory"
)

func main() {
	// --- Example 1: Basic usage with in-memory storage ---
	fmt.Println("=== Example 1: In-Memory Storage ===")

	store := memory.New()
	manager := ocm.NewManager(store)

	// Create a new context
	ctx, err := manager.Create("chat-1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created context: %s\n", ctx.ID)

	// Add messages
	_ = manager.Append("chat-1", ocm.NewMessage(ocm.RoleSystem, "You are a helpful assistant."))
	_ = manager.Append("chat-1", ocm.NewMessage(ocm.RoleUser, "What is Go?"))
	_ = manager.Append("chat-1", ocm.NewMessage(ocm.RoleAssistant, "Go is a programming language."))
	_ = manager.Append("chat-1", ocm.NewMessage(ocm.RoleUser, "Tell me more."))

	// Set custom metadata
	_ = manager.SetMetadata("chat-1", "model", "gpt-4")
	_ = manager.SetMetadata("chat-1", "temperature", 0.7)

	// Retrieve the context
	ctx, _ = manager.Get("chat-1")
	fmt.Printf("Messages: %d\n", ctx.MessageCount())
	for i, msg := range ctx.Messages {
		fmt.Printf("  [%d] %s: %s\n", i, msg.Role, msg.Content)
	}
	fmt.Printf("Metadata: %v\n\n", ctx.Metadata)

	// --- Example 2: With sliding window strategy ---
	fmt.Println("=== Example 2: Sliding Window Strategy ===")

	strategy := ocm.NewSlidingWindowStrategy(3)
	manager2 := ocm.NewManager(store, ocm.WithStrategy(strategy))

	// Get same context but with strategy applied
	ctx, _ = manager2.Get("chat-1")
	fmt.Printf("Messages after strategy (max 3): %d\n", ctx.MessageCount())
	for i, msg := range ctx.Messages {
		fmt.Printf("  [%d] %s: %s\n", i, msg.Role, msg.Content)
	}

	// --- Example 3: Demonstrating storage switchability ---
	fmt.Println("\n=== Example 3: Storage Switchability ===")
	fmt.Println("The Manager accepts any ocm.Storage implementation.")
	fmt.Println("Switch between memory.Store, file.Store, or your own custom store")
	fmt.Println("without changing any business logic code.")

	// List all contexts
	ids, _ := manager.List()
	fmt.Printf("Stored contexts: %v\n", ids)

	// Clean up
	_ = manager.Delete("chat-1")
	fmt.Println("Context deleted.")
}
