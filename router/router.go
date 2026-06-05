package router

import (
	"GIN/handler/handler"
	"GIN/middleware"
	"GIN/repositories"
	services "GIN/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func Init(db *pgxpool.Pool, redis *redis.Client) *gin.Engine {

	services := services.NewServices(db, redis)

	userHandlers := handler.NewUserHandler(services.UserService)
	playlistHandlers := handler.NewPlaylistHandler(services.PlaylistService)
	songsHandlers := handler.NewSongHandler(services.SongService)
	authHandlers := handler.NewAuthHandler(services.AuthService)

	repositories := repositories.NewRepositories(db, redis)
	tokenAuth := middleware.Authorize(repositories)
	cors := middleware.Cors()

	r := gin.Default()
	r.Use(cors)
	userRoutes := r.Group("/users")
	{
		myTracks := userRoutes.Group("/my-tracks")

		userRoutes.Use(tokenAuth)
		myTracks.Use(tokenAuth)

		userRoutes.GET("/:id", userHandlers.GetUser)
		userRoutes.PATCH("/:id", userHandlers.EditUser)
		userRoutes.PATCH("/set-private", userHandlers.EditUserPrivateStatus)
		userRoutes.DELETE("", userHandlers.DeleteUser)
		userRoutes.GET("/my-songs", userHandlers.UserSongs)
		userRoutes.GET("/:id/playlists", userHandlers.GetUserPlaylists)
		userRoutes.GET("/:id/likes", userHandlers.GetUserLikes)

		userRoutes.GET("/listen-statistics", userHandlers.UserListenStatistics)
		userRoutes.GET("/general-listen_statistics", userHandlers.UserGeneralListenStats)
		userRoutes.GET("/likes-count", userHandlers.UserLikesCount)

		myTracks.GET("/listen-statistics", userHandlers.UserTracksListenStatistics)
		myTracks.GET("/likes-count", userHandlers.UserTracksLikesCount)

	}
	songRoutes := r.Group("/songs")
	{
		songRoutes.Use(tokenAuth)

		songRoutes.POST("", songsHandlers.CreateSong)
		songRoutes.DELETE("/:id", songsHandlers.DeleteSong)
		songRoutes.PATCH("/:id", songsHandlers.EditSong)
		songRoutes.GET("/:id", songsHandlers.GetSong)
		songRoutes.GET("/", songsHandlers.GetSongs)
		songRoutes.PATCH("/:id/set-available/:status", songsHandlers.ChangeSongStatus)
		songRoutes.POST("/:id/like", songsHandlers.LikeSong)
		songRoutes.POST("/:id/unlike", songsHandlers.UnLikeSong)
		songRoutes.POST("/:id/add_listen", songsHandlers.AddListen)

	}
	playlistRoutes := r.Group("/playlists")
	{
		playlistRoutes.Use(tokenAuth)

		playlistRoutes.POST("", playlistHandlers.CreatePlaylist)
		playlistRoutes.PATCH("/:id", playlistHandlers.EditPlaylist)
		playlistRoutes.DELETE("/:id", playlistHandlers.DeletePlaylist)
		playlistRoutes.GET("/:id", playlistHandlers.GetPlaylist)
		playlistRoutes.GET("/:id/songs", playlistHandlers.GetPlaylistSongs)

		playlistRoutes.DELETE("/:id/songs/:songID", playlistHandlers.DeleteSongFromPlaylist)
		playlistRoutes.POST("/:id/songs/:songID", playlistHandlers.AddSongToPlaylist)

		playlistRoutes.PATCH("/:id/set-private", playlistHandlers.ChangePlaylistStatus)

		playlistRoutes.POST("/:id/like", playlistHandlers.LikePlaylist)
		playlistRoutes.POST("/:id/unlike", playlistHandlers.UnLikePlaylist)

	}

	r.GET("/refresh", authHandlers.Refresh)
	r.GET("/logout", tokenAuth, authHandlers.Logout)
	r.GET("/logout-all", tokenAuth, authHandlers.LogoutAll)

	r.POST("/registration", authHandlers.Registration)
	r.POST("/login", authHandlers.Login)

	return r
}
