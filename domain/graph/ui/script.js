// Initialize nodes and edges
var isDirected = false
var nodes = new vis.DataSet([
    // {id: 1, label: "1"},
    // {id: 2, label: "2"},
]);

var edges = new vis.DataSet([]);

// Create a graph container
const container = document.getElementById('graph');

// Graph options
const options = {
    layout: {
        improvedLayout: true,
        hierarchical: {
            enabled: false
        }
    },
    edges: {
        color: '#000000' // Edge color
    },
    physics: {
        enabled: true // Enable physics for dynamic layout
    }
};

// Create the network instance
var data = { nodes, edges };
var network = new vis.Network(container, data, options);


// Function to add a new edge from the input form
function addEdge() {
    const fromNode = document.getElementById('fromNode').value;
    const toNode = document.getElementById('toNode').value;
    if (fromNode == toNode || !data.nodes.get(fromNode) || !data.nodes.get(toNode)) {
        alert("please set the id")
        return
    }

    if (!isDirected && edges.get(`${toNode}-${fromNode}`)) {
        alert("can not have duplicate edge in undirected graph")
        return
    }

    w = document.getElementById('weight').value
    if (w == "") {
        w = "1"
    }
    var newEdge = {
        from: fromNode,
        to: toNode,
        id: `${fromNode}-${toNode}`,
        arrows: "to",
        label: w,
        font: {
            size: 16,          // Font size for the label
            face: 'Arial',     // Font family
            color: 'black',    // Font color (black for visibility)
            background: 'white', // Label background color
            align: 'middle',   // Align text to the middle of the edge
            bold: {
                size: 16,        // Bold font size
            },
        },
        color: {
            color: 'black',    // Text color
            background: 'white',  // Background color for the label
            border: 'black',   // Border color for the label
        },
        width: 2,            // Border width of the label
        height: 30
    }
    if (!isDirected) {
        newEdge.arrows = ""
    }
    edges.add(newEdge);
    data = { nodes, edges }
    network.setData(data)
    document.getElementById('fromNode').value = '';
    document.getElementById('toNode').value = '';
    document.getElementById('weight').value = '';
    if (!checkbox.disabled) {
        checkbox.disabled = true

    }

}

var checkbox = document.getElementById("directed")


function addNode() {
    nodeId = document.getElementById("Node").value
    if (nodeId == "" || nodes.get(nodeId)) {
        alert("please set the id")
        return
    }
    nodes.add({ id: nodeId, label: nodeId, color: "lightblue" })
    data = { nodes, edges }
    network.setData(data)
    document.getElementById('Node').value = '';
    if (!checkbox.disabled) {
        checkbox.disabled = true
    }

}

function clearGraph() {
    nodes.clear()
    edges.clear()
    data = { nodes, edges }
    network.setData(data)
    checkbox.disabled = false
}

document.getElementById('Node').addEventListener('keydown', function (event) {
    if (event.key === 'Enter') {
        const nodeId = document.getElementById('Node').value;
        addNode(nodeId);
    }
});

function monitorIsDirected() {
    const checkbox = document.getElementById("directed");

    // Listen for change event
    checkbox.addEventListener("change", function () {
        if (checkbox.checked) {
            isDirected = true
        } else {
            isDirected = false
        }
    });
}

// Call the monitorCheckboxChange function when the page loads
window.onload = monitorIsDirected;

// Get the select element
const algorithmSelect = document.getElementById('algorithmType');
var selectedAlgorithm = "dfs"
// Add an event listener to the select element
algorithmSelect.addEventListener('change', function (event) {
    // Get the selected value

    selectedAlgorithm = event.target.value;
});

const startBox = document.getElementById('start');
var start = ""
startBox.addEventListener("input", function (event) {
    start = event.target.value
})

const endBox = document.getElementById('end');
var end = ""
endBox.addEventListener("input", function (event) {
    end = event.target.value
})

var fileInput = document.getElementById("file-input")
fileInput.addEventListener("change", function (event) {
    clearGraph()
    file = event.target.files[0]
    if (file) {
        // Check if the file is a .txt file (by extension or MIME type)
        const fileExtension = file.name.split('.').pop().toLowerCase();
        const fileType = file.type.toLowerCase();

        // Checking file extension
        if (fileExtension === 'txt' || fileType === 'text/plain') {
            // Use FileReader to read the file content
            const reader = new FileReader();
            reader.onload = function (e) {
                // Read the file
                graph = e.target.result
                lines = graph.split('\n')

                var directed = isDirected
                for (let i = 0; i < lines.length; i++) {
                    var line = lines[i].split(' ')
                    var u = line[0]
                    var v = line[1]
                    var w = "1"
                    if (line.length >= 3) {
                        w = line[2]
                    }
                    if (u == v) {
                        continue
                    }
                    if (!nodes.get(u)) {
                        nodes.add({ id: u, label: u, color: "lightblue" })
                    }
                    if (!nodes.get(v)) {
                        nodes.add({ id: v, label: v, color: "lightblue" })
                    }
                    var newEdge = {
                        from: u,
                        to: v,
                        id: `${u}-${v}`,
                        arrows: "to",
                        label: w,
                        font: {
                            size: 16,          // Font size for the label
                            face: 'Arial',     // Font family
                            color: 'black',    // Font color (black for visibility)
                            background: 'white', // Label background color
                            align: 'middle',   // Align text to the middle of the edge
                            bold: {
                                size: 16,        // Bold font size
                            },
                        },
                        color: {
                            color: 'black',    // Text color
                            background: 'white',  // Background color for the label
                            border: 'black',   // Border color for the label
                        },
                        width: 2,            // Border width of the label
                        height: 30
                    }
                    if (!directed) {
                        newEdge.arrows = ""
                    }
                    if (!edges.get(`${u}-${v}`)) {
                        edges.add(newEdge);
                    }

                }
            };

            // Read the file as text
            reader.readAsText(file);
            checkbox.disabled = true
            fileInput.value = ""

        } else {
            // Not a .txt file, show error message
            errorMessageDiv.textContent = 'Error: Please upload a .txt file.';
        }
    }
})

// const host = process.env.HOST
// const host = "http://0.0.0.0:7000"
const host = "https://api.syamsul.online"
console.log(host);

var log = []
var path = []

function run() {

    if (selectedAlgorithm == "ap" && isDirected) {
        alert("articulation point & bridge for directed graph is not supported")
        return
    }

    var request = {
        // nodes: 
        edges: [],
        nodes: [],
    }

    nodeArr = nodes.get()
    for (let i = 0; i < nodeArr.length; i++) {
        var id = nodeArr[i].id
        request.nodes.push(id)
        nodes.update({ id: id, label: id, color: "lightblue" })
    }

    edgeArr = edges.get()
    for (let i = 0; i < edgeArr.length; i++) {
        var reqEdge = {}
        reqEdge.from = edgeArr[i].from
        reqEdge.to = edgeArr[i].to
        reqEdge.weight = edgeArr[i].label

        request.edges.push(reqEdge)

        var fromNode = reqEdge.from
        var toNode = reqEdge.to
        var w = reqEdge.weight
        var newEdge = {
            from: fromNode,
            to: toNode,
            id: `${fromNode}-${toNode}`,
            arrows: "to",
            label: w,
            font: {
                size: 16,          // Font size for the label
                face: 'Arial',     // Font family
                color: 'black',    // Font color (black for visibility)
                background: 'white', // Label background color
                align: 'middle',   // Align text to the middle of the edge
                bold: {
                    size: 16,        // Bold font size
                },
            },
            color: {
                color: 'black',    // Text color
                background: 'white',  // Background color for the label
                border: 'black',   // Border color for the label
            },
            width: 2,            // Border width of the label
            height: 30
        }
        if (!isDirected) {
            newEdge.arrows = ""
        }
        edges.update(newEdge)
    }


    url = `${host}/graph/${selectedAlgorithm}?start=${start}&end=${end}&isDirected=${isDirected}`

    fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(request)
    }).then(response => {
        if (!response.ok) {
            // If the response is not ok (status code not in the range 200-299), throw an error
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json()
    }).then(async data => {
        log = data.log
        path = data.path

        // await sleep(10000)
        await animate(data.log, data.cycles, data.path, data.acyclic, data.scc, data.ap, data.bridge)
    }).catch(err => {
        console.log(err);
    })
}

const defaultSleep = 500
function sleep() {
    return new Promise(r => setTimeout(r, defaultSleep))
}

async function animate(log, cycles, path, acyclic, scc, ap, bridge) {
    await sleep()
    if (log) {
        for (let i = 0; i < log.length; i++) {
            partition = log[i].split(":")
            type = partition[0]
            switch (type) {
                case "node":
                case "grey":
                case "white":
                case "black":
                    id = partition[1]
                    color = "red"
                    if (type != "node") {
                        color = type
                    }
                    nodes.update({
                        id: id,
                        color: color
                    })
                    break;
                case "edge":
                case "cycle":
                    u = partition[1]
                    v = partition[2]
                    id = `${u}-${v}`
                    if (!edges.get(id)) {
                        id = `${v}-${u}`
                    }
                    color = "blue"
                    if (type == "cycle") {
                        color = "red"
                    }
                    edges.update({
                        id: id,
                        color: {
                            color: color,    // Text color
                            background: 'white',  // Background color for the label
                            border: 'black',   // Border color for the label
                        },
                        width: 10,
                    })
                    break
                case "deNode":
                    continue
                    id = partition[1]
                    nodes.update({
                        id: id,
                        color: "lightblue"
                    })
                    break
                case "deEdge":
                    u = partition[1]
                    v = partition[2]
                    id = `${u}-${v}`
                    if (!edges.get(id)) {
                        id = `${v}-${u}`
                    }
                    edges.update({
                        id: id,
                        color: {
                            color: 'black',    // Text color
                            background: 'white',  // Background color for the label
                            border: 'black',   // Border color for the label
                        },
                        width: 2,
                    })
                    break
                case "bold":
                    u = partition[1]
                    nodes.update({
                        id: u,
                        borderWidth: 10,
                    })
                    break
                case "deBold":
                    continue
                    var u = partition[1]
                    nodes.update({
                        id: u,
                        borderWidth: 1,
                    })
                    break

                case "label":
                    var id = partition[1]
                    var tin = partition[2]
                    var low = partition[3]

                    nodes.update({
                        id: id,
                        label: `${id},T=${tin},low=${low}`
                    })
                    break
                case "ap":
                    var id = partition[1]
                    nodes.update({
                        id: id,
                        color: "red",
                        borderWidth: 10,
                    })
                    break
                case "bridge":
                    var u = partition[1]
                    var v = partition[2]
                    var edgeId = `${u}-${v}`
                    if (!edges.get(edgeId)) {
                        edgeId = `${v}-${u}`
                    }

                    edges.update({
                        id: edgeId,
                        width: 10,
                        color: "red"
                    })
                    break
                default:
                    break;
            }

            await sleep()
        }
    }
    if (cycles) {
        for (let i = 0; i < cycles.length; i++) {
            for (let j = 0; j < cycles[i].length; j++) {
                a = j
                b = (j + 1) % cycles[i].length
                id = `${cycles[i][j]}-${cycles[i][b]}`
                if (!edges.get(id)) {
                    id = `${cycles[i][b]}-${cycles[i][j]}`
                }

                edges.update({
                    id: id,
                    color: {
                        color: "red",    // Text color
                        background: 'white',  // Background color for the label
                        border: 'black',   // Border color for the label
                    },
                    width: 10,
                })
                await sleep()
            }
        }
    }

    if (selectedAlgorithm == "dag") {
        if (!acyclic) {
            alert("graph is non acyclic")
        } else {
            for (let i = 0; i < path.length; i++) {
                id = path[i]
                nodes.update({
                    id: id,
                    label: `${i + 1}(id:${id})`,
                    borderWidth: 15,
                })

                await sleep()
            }
        }
    }

    if (selectedAlgorithm == "scc") {
        var colorPool = ["white", "green", "lightblue", "red", "orange", "black"]
        for (let i = 0; i < scc.length; i++) {
            var group = scc[i]
            for (let j = 0; j < group.length; j++) {
                var node = group[j]
                nodes.update({
                    id: node,
                    color: colorPool[i % colorPool.length]
                })
                await sleep()
            }
        }
    }

    if (selectedAlgorithm == "ap") {
        for (let i = 0; i < ap.length; i++) {
            var id = ap[i]
            nodes.update({
                id: id,
                color: "red",
                borderWidth: 10,
            })
        }

        for (let i = 0; i < bridge.length; i++) {
            var u = bridge[i][0]
            var v = bridge[i][1]
            var edgeId = `${u}-${v}`
            if (!edges.get(edgeId)) {
                edgeId = `${v}-${u}`
            }

            edges.update({
                id: edgeId,
                width: 10,
                color: "red"
            })
        }
    }

    if (selectedAlgorithm == "ep"){
        if (path.length == 0) {
            alert("graph doesn't have eulerian path/circut")
            return
        }

        for (let i=0; i<path.length-1; i++) {
            var u = path[i]
            var v = path[i+1]

            var id = `${u}-${v}`
            var arrows = "to"
            if (!edges.get(id)) {
                id = `${v}-${u}`
                arrows = "from"
            }

            edges.update({
                id: id,
                width: 8,
                color: "red",
                label: `${i+1}`,
                arrows: arrows,
            })

            await sleep()
        }
    }
}