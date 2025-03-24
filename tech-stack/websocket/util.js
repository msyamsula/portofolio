
export function createRoomName(userId, pairId) {
    userId = parseInt(userId)
    pairId = parseInt(pairId)
    let room = `${Math.min(userId,pairId)},${Math.max(userId,pairId)}`
    return room
}


// const lookupd = "http://0.0.0.0:4161"
const lookupd = "http://nsqlookupd-clusterip:4161"
export async function getNsqd(){

    let response = await fetch(`${lookupd}/nodes`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    })

    let data = await response.json()
    if (data) {
        return data
    }

    return null
}