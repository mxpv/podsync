# How to get YouTube API Key

1. Navigate to https://console.developers.google.com
2. Click `Select a project`.
![Select project](img/youtube_select_project.png)
3. Click `New project`.
![New project](img/youtube_new_project.png)
4. Give it a name and click `Create` button.
![Dashboard](img/youtube_dashboard.png)
5. Click `Library`, find and click on `YouTube Data API v3` box.
![YouTube Data API](img/youtube_data_api_v3.png)
6. Click `Enable`.
![YouTube Enable](img/youtube_data_api_enable.png)
5. Click `Credentials`.
6. Click `Create credentials`.
7. Select `API key`.
![Create API key](img/youtube_create_api_key.png)
8. Copy token to your CLI's configuration file or set it as an environment variable:
![Copy token](img/youtube_copy_token.png)
```toml
[tokens]
youtube = "AIzaSyD4w2s-k79YNR98ABC"
```
Or set the environment variable:
```sh
export PODSYNC_YOUTUBE_API_KEY="AIzaSyD4w2s-k79YNR98ABC"
```

For API key rotation, you can specify multiple keys separated by spaces:
```sh
export PODSYNC_YOUTUBE_API_KEY="AIzaSyD4w2s-k79YNR98ABC AIzaSyD4w2s-k79YNR98DEF"
```