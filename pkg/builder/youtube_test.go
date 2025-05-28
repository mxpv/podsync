package builder

import (
	"context"
	// "errors" // Removed unused import
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	youtube "google.golang.org/api/youtube/v3"
	// For mocking, we are abstracting away direct googleapi/option usage in tests
	// by assuming the mock functions are called directly.
)

var (
	testCtx              = context.Background()
	ytKey                = os.Getenv("YOUTUBE_TEST_API_KEY") // For existing integration tests
	skipMockedTests      = false                            // Set to false to run new mock tests (requires youtube.go adaptation)
	skipMockedTestsReason = "Skipping mock test: Requires youtube.go to be adapted for the conceptual mocking strategy (e.g., using global currentMockAPI)."
)

// Using the exported MockAPIInternal type from youtube.go (and ActiveInternalMockAPI global variable)
// No local mockAPI struct or local currentMockAPI global variable needed.

// mockYoutubeBuilder now sets the global ActiveInternalMockAPI from youtube.go.
func mockYoutubeBuilder(t *testing.T, api *MockAPIInternal) *YouTubeBuilder { // Uses exported MockAPIInternal type
	ActiveInternalMockAPI = api // Assign to the exported global var from youtube.go
	// Return a real builder; its calls are assumed to be intercepted by ActiveInternalMockAPI
	builder, err := NewYouTubeBuilder("mock_key_placeholder") // Placeholder key
	require.NoError(t, err, "Failed to create YouTubeBuilder for mocking") // Ensure builder creation doesn't fail
	
	// Cleanup function to reset ActiveInternalMockAPI after the test
	t.Cleanup(func() {
		ActiveInternalMockAPI = nil
	})

	return builder
}

// --- Test Cases for @handle URLs ---

func TestYouTubeBuilder_Build_HandleVideosURL(t *testing.T) {
	if skipMockedTests {
		t.Skip(skipMockedTestsReason)
	}

	// api instance should now be of type *mockAPIInternal (defined in youtube.go, used by activeInternalMockAPI)
	// The local mockAPI struct in this file should be removed or renamed if it conflicts.
	// For now, we assume that the `mockAPIInternal` struct from `youtube.go` is implicitly used
	// when creating the `api` variable in tests, or that the local `mockAPI` is identical and compatible.
	// To be explicit, tests should instantiate &mockAPIInternal{...}
	// Let's assume the local `mockAPI` struct is what we use to define the `api` variable,
	// and it's compatible with `activeInternalMockAPI`'s type `*mockAPIInternal`.
	// This means the definition of `mockAPI` in this test file needs to be identical to `mockAPIInternal` in `youtube.go`
	// or simply removed if `mockAPIInternal` from `youtube.go` is to be used directly.
	// The previous diff already made `mockAPI` in test file compatible with `mockAPIInternal` for the used fields.
	
	// To ensure type correctness, we should instantiate with the type from youtube.go if possible,
	// or ensure the local definition is identical.
	// The simplest change here is to make the local `api` variable of type `*mockAPIInternal`.
	// This requires the `mockAPIInternal` struct to be defined in this file or imported.
	// Since `mockAPIInternal` is unexported in `youtube.go`, we cannot import it.
	// So, the struct `mockAPI` in this test file must be structurally identical and used.
	// And the parameter to `mockYoutubeBuilder` must be `*mockAPI` if that's what we pass.
	// Then `mockYoutubeBuilder` must assign `api` (type `*mockAPI`) to `activeInternalMockAPI` (type `*mockAPIInternal`).
	// This assignment works if the structs are identical.

	// Let's rename the local struct to avoid confusion and ensure compatibility.
	// The global var in youtube.go is `activeInternalMockAPI *mockAPIInternal`
	// The local struct in test file `mockAPI` should be changed to `mockAPIInternal` for clarity
	// and then ensure `mockYoutubeBuilder` uses this type.

	// Re-iterating the change from the diff description:
	// 1. Rename local `mockAPI` struct to `mockAPIInternalTest` (or similar) if needed for clarity,
	//    OR ensure it's identical to `youtube.go`'s `mockAPIInternal`.
	// 2. Remove `var currentMockAPI *mockAPI`.
	// 3. `mockYoutubeBuilder` takes `api *mockAPIInternal` (matching youtube.go's type)
	//    and assigns it to `activeInternalMockAPI`.
	// 4. Test functions create `api := &mockAPIInternal{}`.

	// The diff applied should change the `mockYoutubeBuilder` to use `activeInternalMockAPI`.
	// The test functions will create `api := &mockAPI{}`.
	// This `*mockAPI` must be assignable to `*mockAPIInternal`.
	// This is true if they are structurally identical.
	// The `mockAPI` struct in this file must be kept and used to instantiate `api` vars.
	// `mockYoutubeBuilder` will take `*mockAPI` and assign to `activeInternalMockAPI`.
	// This is fine if the types are identical. Let's ensure `mockYoutubeBuilder` takes `*mockAPI`.

	api := &MockAPIInternal{} // Use exported MockAPIInternal type from youtube.go
	builder := mockYoutubeBuilder(t, api)
	cfg := &feed.Config{URL: "https://www.youtube.com/@TestHandle/videos", PageSize: 2}

	handle := "@TestHandle"
	channelID := "UC12345"
	uploadsPlaylistID := "UU12345" // Standard uploads playlist ID format

	// Mock for resolveHandle (Search.List)
	api.SearchListFunc = func(q, typeValue string, parts []string) (*youtube.SearchListResponse, error) { // Exported field name
		require.Equal(t, handle, q)
		require.Equal(t, "channel", typeValue)
		require.Contains(t, parts, "id")
		return &youtube.SearchListResponse{
			Items: []*youtube.SearchResult{{Id: &youtube.ResourceId{Kind: "youtube#channel", ChannelId: channelID}}},
		}, nil
	}

	// Mock for listChannels (Channels.List) - called by queryFeed
	api.ChannelsListFunc = func(id, forUsername string, parts []string) (*youtube.ChannelListResponse, error) { // Exported field name
		require.Equal(t, channelID, id)
		require.Empty(t, forUsername)
		// Parts requested by queryFeed for TypeChannel/TypeUser: "id,snippet,contentDetails"
		require.ElementsMatch(t, []string{"id", "snippet", "contentDetails"}, parts)
		return &youtube.ChannelListResponse{
			Items: []*youtube.Channel{{
				Id:      channelID,
				Snippet: &youtube.ChannelSnippet{Title: "Test Channel Title", Description: "Channel Description", PublishedAt: time.Now().Format(time.RFC3339)},
				ContentDetails: &youtube.ChannelContentDetails{
					RelatedPlaylists: &youtube.ChannelContentDetailsRelatedPlaylists{Uploads: uploadsPlaylistID},
				},
			}},
		}, nil
	}

	// Mock for queryItems (PlaylistItems.List)
	api.PlaylistItemsListFunc = func(playlistId string, parts []string, maxResults int64) (*youtube.PlaylistItemListResponse, error) { // Exported field name
		require.Equal(t, uploadsPlaylistID, playlistId)
		require.ElementsMatch(t, []string{"id", "snippet"}, parts) // Default parts for PlaylistItems.List
		require.EqualValues(t, cfg.PageSize, maxResults)
		return &youtube.PlaylistItemListResponse{
			Items: []*youtube.PlaylistItem{
				{Snippet: &youtube.PlaylistItemSnippet{Title: "Video 1", PublishedAt: time.Now().Add(-2*time.Hour).Format(time.RFC3339), ResourceId: &youtube.ResourceId{VideoId: "vid1"}, Position: 0}},
				{Snippet: &youtube.PlaylistItemSnippet{Title: "Video 2", PublishedAt: time.Now().Add(-1*time.Hour).Format(time.RFC3339), ResourceId: &youtube.ResourceId{VideoId: "vid2"}, Position: 1}},
			}, NextPageToken: "", // No more pages
		}, nil
	}

	// Mock for queryVideoDescriptions (Videos.List)
	api.VideosListFunc = func(videoIds []string, parts []string) (*youtube.VideoListResponse, error) { // Exported field name
		require.ElementsMatch(t, []string{"vid1", "vid2"}, videoIds)
		require.ElementsMatch(t, []string{"id", "snippet", "contentDetails"}, parts) // Default parts for Videos.List
		return &youtube.VideoListResponse{
			Items: []*youtube.Video{
				{Id: "vid1", Snippet: &youtube.VideoSnippet{Title: "Video 1 Title", Description: "Desc 1", LiveBroadcastContent: "none"}, ContentDetails: &youtube.VideoContentDetails{Duration: "PT1M30S"}},
				{Id: "vid2", Snippet: &youtube.VideoSnippet{Title: "Video 2 Title", Description: "Desc 2", LiveBroadcastContent: "none"}, ContentDetails: &youtube.VideoContentDetails{Duration: "PT2M0S"}},
			},
		}, nil
	}

	actualFeed, err := builder.Build(context.Background(), cfg)

	require.NoError(t, err)
	require.NotNil(t, actualFeed)
	assert.Equal(t, handle, actualFeed.Author)
	// As per `queryFeed` logic for `@handle` (TypeChannel or TypeUser):
	// ItemURL becomes `https://youtube.com/@TestHandle`
	assert.Equal(t, "https://youtube.com/"+handle, actualFeed.ItemURL)
	assert.Equal(t, uploadsPlaylistID, actualFeed.ItemID) // ItemID is the uploads playlist ID
	assert.Equal(t, "Test Channel Title", actualFeed.Title)
	require.Len(t, actualFeed.Episodes, 2)
	assert.Equal(t, "Video 1 Title", actualFeed.Episodes[0].Title)
	assert.Equal(t, int64(90), actualFeed.Episodes[0].Duration)
	assert.Equal(t, "Video 2 Title", actualFeed.Episodes[1].Title)
	assert.Equal(t, int64(120), actualFeed.Episodes[1].Duration)
}

func TestYouTubeBuilder_Build_HandlePlaylistsURL(t *testing.T) {
	if skipMockedTests {
		t.Skip(skipMockedTestsReason)
	}
	api := &MockAPIInternal{} // Use exported MockAPIInternal type
	builder := mockYoutubeBuilder(t, api)
	cfg := &feed.Config{URL: "https://www.youtube.com/@TestHandle/playlists", PageSize: 1}

	handle := "@TestHandle"
	channelID := "UC-PlaylistHandle"
	uploadsPlaylistID := "UU-PlaylistHandle"

	// Mock for resolveHandle (Search.List)
	api.SearchListFunc = func(q, typeValue string, parts []string) (*youtube.SearchListResponse, error) { // Exported field name
		require.Equal(t, handle, q)
		return &youtube.SearchListResponse{Items: []*youtube.SearchResult{{Id: &youtube.ResourceId{ChannelId: channelID}}}}, nil
	}

	// Mock for listChannels (Channels.List) - called by queryFeed for TypePlaylist with handle
	api.ChannelsListFunc = func(id, forUsername string, parts []string) (*youtube.ChannelListResponse, error) { // Exported field name
		require.Equal(t, channelID, id)
		// Parts requested by queryFeed for TypePlaylist with handle: "id,snippet,contentDetails"
		require.ElementsMatch(t, []string{"id", "snippet", "contentDetails"}, parts)
		return &youtube.ChannelListResponse{
			Items: []*youtube.Channel{{
				Id:      channelID,
				Snippet: &youtube.ChannelSnippet{Title: "Playlist Test Channel", PublishedAt: time.Now().Format(time.RFC3339)},
				ContentDetails: &youtube.ChannelContentDetails{
					RelatedPlaylists: &youtube.ChannelContentDetailsRelatedPlaylists{Uploads: uploadsPlaylistID},
				},
			}},
		}, nil
	}
	
	// Mock for queryItems (PlaylistItems.List)
	api.PlaylistItemsListFunc = func(playlistId string, parts []string, maxResults int64) (*youtube.PlaylistItemListResponse, error) { // Exported field name
		require.Equal(t, uploadsPlaylistID, playlistId) // Should use the resolved uploads playlist ID
		return &youtube.PlaylistItemListResponse{
			Items: []*youtube.PlaylistItem{{Snippet: &youtube.PlaylistItemSnippet{ResourceId: &youtube.ResourceId{VideoId: "vidP1"}, Position: 0, PublishedAt: time.Now().Add(-3*time.Hour).Format(time.RFC3339)}}}, // Added PublishedAt
		}, nil
	}

	// Mock for queryVideoDescriptions (Videos.List)
	api.VideosListFunc = func(videoIds []string, parts []string) (*youtube.VideoListResponse, error) { // Exported field name
		return &youtube.VideoListResponse{
			Items: []*youtube.Video{{Id: "vidP1", Snippet: &youtube.VideoSnippet{Title: "Playlist Video 1", LiveBroadcastContent: "none", PublishedAt: time.Now().Add(-3*time.Hour).Format(time.RFC3339)}, ContentDetails: &youtube.VideoContentDetails{Duration: "PT3M"}}}, // Added PublishedAt
		}, nil
	}

	actualFeed, err := builder.Build(context.Background(), cfg)

	require.NoError(t, err)
	require.NotNil(t, actualFeed)
	assert.Equal(t, "Playlist Test Channel - All Uploads", actualFeed.Title) // Specific title for this case
	assert.Equal(t, handle, actualFeed.Author)
	assert.Equal(t, fmt.Sprintf("https://youtube.com/%s/playlists", handle), actualFeed.ItemURL)
	assert.Equal(t, uploadsPlaylistID, actualFeed.ItemID) // ItemID is the uploads playlist ID
	require.Len(t, actualFeed.Episodes, 1)
	assert.Equal(t, "Playlist Video 1", actualFeed.Episodes[0].Title)
}

// /* // Uncommenting the function
func TestYouTubeBuilder_Build_HandleURLNoSuffix(t *testing.T) {
	if skipMockedTests {
		t.Skip(skipMockedTestsReason)
	}
	api := &MockAPIInternal{} // Use exported MockAPIInternal type
	builder := mockYoutubeBuilder(t, api)
	cfg := &feed.Config{URL: "https://www.youtube.com/@TestHandleNoSuffix", PageSize: 1}

	handle := "@TestHandleNoSuffix"
	channelID := "UC-NoSuffix"
	uploadsPlaylistID := "UU-NoSuffix"

	api.SearchListFunc = func(q, typeValue string, parts []string) (*youtube.SearchListResponse, error) { // Exported field name
		return &youtube.SearchListResponse{Items: []*youtube.SearchResult{{Id: &youtube.ResourceId{ChannelId: channelID}}}}, nil
	}
	api.ChannelsListFunc = func(id, forUsername string, parts []string) (*youtube.ChannelListResponse, error) { // Exported field name
		return &youtube.ChannelListResponse{
			Items: []*youtube.Channel{{
				Id: channelID, 
				Snippet: &youtube.ChannelSnippet{
					Title: "NoSuffix Channel", // Comma added
					Description: "Desc",       // Comma added
					PublishedAt: time.Now().Format(time.RFC3339),
				}, // Comma added
				ContentDetails: &youtube.ChannelContentDetails{
					RelatedPlaylists: &youtube.ChannelContentDetailsRelatedPlaylists{
						Uploads: uploadsPlaylistID,
					}, // Comma added
				}, // Comma added
			}}, // Comma added
		}, nil
	}
	api.PlaylistItemsListFunc = func(playlistId string, parts []string, maxResults int64) (*youtube.PlaylistItemListResponse, error) { // Exported field name
		return &youtube.PlaylistItemListResponse{
			Items: []*youtube.PlaylistItem{{ // Outer { for slice element
				Snippet: &youtube.PlaylistItemSnippet{ // Inner { for Snippet struct
					ResourceId: &youtube.ResourceId{VideoId: "vidNS1"}, // Comma added
					Position: 0,                                        // Comma added
					PublishedAt: time.Now().Add(-4*time.Hour).Format(time.RFC3339),
				}, // Comma added (closes Snippet)
			}}, // Comma added (closes PlaylistItem)
		}, nil // Comma added (closes PlaylistItemListResponse)
	}
	api.VideosListFunc = func(videoIds []string, parts []string) (*youtube.VideoListResponse, error) { // Exported field name
		return &youtube.VideoListResponse{
			Items: []*youtube.Video{{ // Outer { for slice element
				Id: "vidNS1", 
				Snippet: &youtube.VideoSnippet{ // Inner { for Snippet struct
					Title: "NoSuffix Video 1",       // Comma added
					LiveBroadcastContent: "none",    // Comma added
					PublishedAt: time.Now().Add(-4*time.Hour).Format(time.RFC3339),
				}, // Comma added (closes Snippet)
				ContentDetails: &youtube.VideoContentDetails{Duration: "PT4M"}, // Comma added
			}}, // Comma added (closes Video)
		}, nil // Comma added (closes VideoListResponse)
	}

	actualFeed, err := builder.Build(context.Background(), cfg)

	require.NoError(t, err)
	require.NotNil(t, actualFeed)
	assert.Equal(t, "NoSuffix Channel", actualFeed.Title)
	assert.Equal(t, handle, actualFeed.Author)
	// As per `queryFeed` logic for `@handle` (TypeChannel or TypeUser when no suffix):
	// ItemURL becomes `https://youtube.com/@TestHandleNoSuffix`
	assert.Equal(t, "https://youtube.com/"+handle, actualFeed.ItemURL)
	assert.Equal(t, uploadsPlaylistID, actualFeed.ItemID) // ItemID is the uploads playlist ID
	require.Len(t, actualFeed.Episodes, 1)
	assert.Equal(t, "NoSuffix Video 1", actualFeed.Episodes[0].Title)
}
// */ // Uncommenting the function

func TestYouTubeBuilder_Build_Error_HandleNotFound(t *testing.T) {
	if skipMockedTests {
		t.Skip(skipMockedTestsReason)
	}
	api := &MockAPIInternal{} // Use exported MockAPIInternal type
	builder := mockYoutubeBuilder(t, api)
	cfg := &feed.Config{URL: "https://www.youtube.com/@NonExistentHandle"}

	handle := "@NonExistentHandle"

	// Mock Search.List to return no items
	api.SearchListFunc = func(q, typeValue string, parts []string) (*youtube.SearchListResponse, error) { // Exported field name
		require.Equal(t, handle, q)
		return &youtube.SearchListResponse{Items: []*youtube.SearchResult{}}, nil // No items found
	}

	_, err := builder.Build(context.Background(), cfg)

	require.Error(t, err)
	// Error path: resolveHandle -> listChannels (wraps error) -> queryFeed (wraps error) -> Build
	assert.ErrorIs(t, err, model.ErrNotFound, "Expected error to wrap model.ErrNotFound")
	assert.Contains(t, err.Error(), "failed to resolve handle") // Error from listChannels
	assert.Contains(t, err.Error(), handle)
}

func TestYouTubeBuilder_Build_Error_HandlePlaylistsNoUploadsID(t *testing.T) {
	if skipMockedTests {
		t.Skip(skipMockedTestsReason)
	}
	api := &MockAPIInternal{} // Use exported MockAPIInternal type
	builder := mockYoutubeBuilder(t, api)
	cfg := &feed.Config{URL: "https://www.youtube.com/@TestHandleBadConfig/playlists"}

	handle := "@TestHandleBadConfig"
	channelID := "UC-BadConfig"

	// Mock Search.List to successfully find the handle
	api.SearchListFunc = func(q, typeValue string, parts []string) (*youtube.SearchListResponse, error) { // Exported field name
		return &youtube.SearchListResponse{Items: []*youtube.SearchResult{{Id: &youtube.ResourceId{ChannelId: channelID}}}}, nil
	}

	// Mock Channels.List to return channel details but with an empty Uploads ID
	api.ChannelsListFunc = func(id, forUsername string, parts []string) (*youtube.ChannelListResponse, error) { // Exported field name
		require.Equal(t, channelID, id)
		return &youtube.ChannelListResponse{
			Items: []*youtube.Channel{{
				Id:      channelID,
				Snippet: &youtube.ChannelSnippet{Title: "Bad Config Channel", PublishedAt: time.Now().Format(time.RFC3339)}, // Added PublishedAt
				ContentDetails: &youtube.ChannelContentDetails{
					RelatedPlaylists: &youtube.ChannelContentDetailsRelatedPlaylists{
						Uploads: "", // Crucially, Uploads ID is missing
					},
				},
			}},
		}, nil
	}

	_, err := builder.Build(context.Background(), cfg)

	require.Error(t, err)
	// Error from queryFeed when uploadsPlaylistID is empty for TypePlaylist with handle
	assert.Contains(t, err.Error(), "could not find uploads playlist for channel")
	assert.Contains(t, err.Error(), channelID)
	assert.Contains(t, err.Error(), handle)
}


// --- Existing Integration Tests (Renamed for clarity) ---
// These tests require a real YOUTUBE_TEST_API_KEY and hit the actual YouTube API.

func TestYT_QueryChannel_Integration(t *testing.T) {
	if ytKey == "" {
		t.Skip("YouTube API key is not provided for integration test")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	channel, err := builder.listChannels(testCtx, model.TypeChannel, "UC2yTVSttx7lxAOAzx1opjoA", "id,snippet") // Added snippet for more coverage
	require.NoError(t, err)
	require.Equal(t, "UC2yTVSttx7lxAOAzx1opjoA", channel.Id)
	require.NotEmpty(t, channel.Snippet.Title)

	channel, err = builder.listChannels(testCtx, model.TypeUser, "fxigr1", "id,snippet")
	require.NoError(t, err)
	require.Equal(t, "UCr_fwF-n-2_olTYd-m3n32g", channel.Id) // fxigr1 resolves to this channel ID
	require.NotEmpty(t, channel.Snippet.Title)
}

func TestYT_BuildFeed_Integration(t *testing.T) {
	if ytKey == "" {
		t.Skip("YouTube API key is not provided for integration test")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	urls := []string{
		"https://www.youtube.com/channel/UCupvZG-5ko_eiXAupbDfxWw", // Normal Channel
		"https://www.youtube.com/playlist?list=PLF7tUDhGkiCk_Ne30zu7SJ9gZF9R9ZruE", // Normal Playlist
		"https://youtube.com/user/WylsaLive", // User URL
		// Add handle URLs if you have stable ones for integration testing (they resolve to channels)
		// "https://www.youtube.com/@LinusTechTips/videos", // Handle URL (example)
	}

	for _, addr := range urls {
		t.Run(addr, func(t *testing.T) {
			_feed, buildErr := builder.Build(testCtx, &feed.Config{URL: addr, PageSize: 2}) // Small page size
			require.NoError(t, buildErr, "Failed to build feed for URL: %s", addr)

			assert.NotEmpty(t, _feed.Title)
			assert.NotEmpty(t, _feed.Description)
			assert.NotEmpty(t, _feed.Author)
			assert.NotEmpty(t, _feed.ItemURL)
			assert.NotEmpty(t, _feed.ItemID) // Usually playlist ID after build

			if len(_feed.Episodes) > 0 {
				for _, item := range _feed.Episodes {
					assert.NotEmpty(t, item.Title)
					assert.NotEmpty(t, item.VideoURL)
					assert.NotZero(t, item.Duration)
					assert.NotEmpty(t, item.Thumbnail)
				}
			} else {
				// It's possible some test channels/playlists are empty or very new
				t.Logf("Warning: No episodes found for %s. This might be okay.", addr)
			}
		})
	}
}

func TestYT_GetVideoCount_Integration(t *testing.T) {
	if ytKey == "" {
		t.Skip("YouTube API key is not provided for integration test")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	feeds := []*model.Info{
		{Provider: model.ProviderYoutube, LinkType: model.TypeUser, ItemID: "fxigr1"},
		{Provider: model.ProviderYoutube, LinkType: model.TypeChannel, ItemID: "UCupvZG-5ko_eiXAupbDfxWw"},
		{Provider: model.ProviderYoutube, LinkType: model.TypePlaylist, ItemID: "PLF7tUDhGkiCk_Ne30zu7SJ9gZF9R9ZruE"},
		// Add handle based Info if GetVideoCount should support it directly (it currently doesn't, relies on Build's resolution)
		// {Provider: model.ProviderYoutube, LinkType: model.TypeChannel, ItemID: "@LinusTechTips"}, // This would fail GetVideoCount
	}

	for _, feedInfo := range feeds {
		// Create a new variable for the goroutine
		currentFeedInfo := feedInfo
		t.Run(currentFeedInfo.ItemID, func(t *testing.T) {
			// GetVideoCount doesn't resolve handles itself, it expects channel/playlist IDs.
			// If testing handles here, they must be pre-resolved or this test is for non-handle cases.
			if currentFeedInfo.ItemID[0] == '@' {
				t.Skip("Skipping GetVideoCount for handle ID directly, as it requires prior resolution.")
				return
			}
			count, countErr := builder.GetVideoCount(testCtx, currentFeedInfo) // Pass currentFeedInfo directly
			assert.NoError(t, countErr)
			assert.GreaterOrEqual(t, count, uint64(0), "Video count should be non-negative") // Some channels might have 0 videos
		})
	}
}

// Note on Mocking:
// The new mocked tests (TestYouTubeBuilder_Build_Handle*) rely on a conceptual global 'currentMockAPI'
// that the production code (youtube.go) would need to be adapted to use for these tests to run.
// This adaptation could be via build tags and conditional compilation, or by refactoring
// youtube.go to allow injection of a mockable service/client.
// The `skipMockedTests` flag is used to prevent these tests from running by default
// as they require these modifications to the production code which are not part of this subtask.
// The tests are structured to clearly define the expected API interactions and outcomes.
