package persistent

var queryGetLongUrl = "SELECT long_url FROM url WHERE short_url = :short_url"
var queryGetShortUrl = "SELECT short_url FROM url WHERE long_url = :long_url"
var querySetShortUrl = "INSERT INTO url (short_url, long_url) VALUES (:short_url, :long_url)"
