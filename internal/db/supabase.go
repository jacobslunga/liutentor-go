package db

import "github.com/supabase-community/supabase-go"

func NewSupabaseClient(url, key string) (*supabase.Client, error) {
	return supabase.NewClient(url, key, nil)
}
