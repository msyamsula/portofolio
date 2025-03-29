// const apiHost = "http://0.0.0.0:12000"

const apiHost = "https://api.syamsul.online"


async function getFriends(id) {
    let response = await fetch(`${apiHost}/user/friend?id=${id}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    })

    let user = await response.json()
    if (user.data) {
        return user.data
    }

    return null
}

async function connectWithUser(idA, idB) {
    idA = parseInt(idA, 10)
    let edge = {
        small_id: idA,
        big_id: idB
    }
    let response = await fetch(`${apiHost}/user/friend`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(edge)
    })

    let data = await response.json()
    if (data) {
        return data
    }
    return null
}

async function getUser(username) {
    const query = new URLSearchParams({
        username: username
    }).toString()
    response = await fetch(`${apiHost}/user?${query}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    let user = await response.json()
    if (!user.error) {
        return user.data
    }

    return null

}

async function registerUser(username, online) {
    try {
        let response = await fetch(`${apiHost}/user`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username,
                online: online,
            })
        })

        let user = await response.json()
        if (!user.error) {
            return user.data
        }

        return null
    } catch (error) {
        console.log(error, "goes here");
        return null
    }
}

async function getConversation(userId, pairId) {
    const query = new URLSearchParams({
        userId: userId,
        pairId: pairId,
    }).toString()
    response = await fetch(`${apiHost}/message?${query}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    let conversation = await response.json()
    if (!conversation.error) {
        return conversation.data
    }

    return null

}