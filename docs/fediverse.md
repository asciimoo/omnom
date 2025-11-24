# Fediverse Support

Omnom has entered the Fediverse with the implementation of ActivityPub support, enabling seamless communication with platforms like Mastodon.

This integration means Omnom users can be followed from other Fediverse platforms, ensuring that important saved websites are never missed by their followers.

Furthermore Omnom allows you to follow actors across the Fediverse. This means you can add the handles of users from platforms like Mastodon, Pleroma, or other compatible services directly into your Omnom feeds.
The public posts and activities from the Fediverse actors you follow will be fetched and integrated into your main content feed within Omnom. This allows you to monitor updates from the decentralized web alongside your RSS/Atom feeds and saved bookmarks.

## Usage

Every Omnom user is a valid ActivityPub [actor](https://www.w3.org/TR/activitypub/#actors) which can be referenced by either `[username]@[omnom.domain]` (e.g. `testuser@omnom.zone`) or using the URL of their profile page (e.g. `https://omnom.zone/users/testuser`). Use one of these user handles in other Fediverse platforms to allow those services to discover Omnom users and make following available.

### Following others from Omnom

Navigate to your [feeds](feeds) page where you can input the ActivityPub profile URLs of individuals or services you wish to follow by adding a new feed in the left sidebar.

### Follow you from other services

Example in Mastodon:
![Mastodon follow](/static/images/docs/omnom_mastodon_follow.png)

The only thing left to do is hitting the "follow" button, after the Omnom user has been successfully found. Every new bookmark created from that point will show up in your feed.

Example in mastodon:
![Mastodon follow](/static/images/docs/omnom_mastodon_post.png)
