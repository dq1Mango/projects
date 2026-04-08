function cubicBezier(x1, y1, x2, y2) {
  const Ay = 1 - 3 * y2 + 3 * y1;
  const By = 3 * y2 - 6 * y1;
  const Cy = 3 * y1;

  return function (progress) {
    return ((Ay * progress + By) * progress + Cy) * progress;
  };
}

const easeInOut = cubicBezier(0.65, 0, 0.35, 1);
const linear = cubicBezier(0, 0, 1, 1);
const easeOut = cubicBezier(0, 0, 0.58, 1);
const easeIn = cubicBezier(0.32, 0, 0.67, 0);

const flowLength = 30; // px

class MessageGraphVisualization {
  constructor() {
    this.canvas = document.getElementById("graphCanvas");
    this.ctx = this.canvas.getContext("2d");

    // Graph data
    this.nodes = [];
    this.edges = [];

    // Simulation parameters (will be read from DOM)
    this.repulsionForce = null;
    this.linkDistance = null;
    this.springConstant = null;
    this.damping = 0.9;

    // View parameters (will be read from DOM)
    this.showLabels = null;
    this.showEdgeWeights = null;
    this.zoom = 1;
    this.offsetX = 0;
    this.offsetY = 0;

    // Message flow animation parameters (will be read from DOM)
    this.animationDuration = null;
    this.animationIntensity = null;
    this.autoFlow = null;
    this.messageFlows = [];
    this.lastAutoFlowTime = 0;

    // Interaction state
    this.isDragging = false;
    this.dragNode = null;
    this.lastMouseX = 0;
    this.lastMouseY = 0;

    this.init();
  }

  init() {
    this.setupCanvas();
    this.readInitialValues();
    this.setupEventListeners();
    this.generateSampleData();
    this.startAnimation();
  }

  readInitialValues() {
    // Read simulation parameters from sliders
    this.repulsionForce = parseInt(document.getElementById("repulsionSlider").value);
    this.linkDistance = parseInt(document.getElementById("linkDistanceSlider").value);
    this.springConstant = parseFloat(document.getElementById("springConstantSlider").value);

    // Read view parameters from checkboxes
    this.showLabels = document.getElementById("showLabels").checked;
    this.showEdgeWeights = document.getElementById("showEdgeWeights").checked;

    // Read animation parameters from sliders and checkboxes
    this.animationDuration = parseInt(document.getElementById("animationDurationSlider").value);
    this.animationIntensity = parseInt(document.getElementById("animationIntensitySlider").value);
    this.autoFlow = document.getElementById("autoFlow").checked;

    // Update labels with initial values
    this.updateSliderLabels();
  }

  updateSliderLabels() {
    document.getElementById("repulsionLabel").textContent = `Node Repulsion: ${this.repulsionForce}`;
    document.getElementById("linkDistanceLabel").textContent = `Link Distance: ${this.linkDistance}px`;
    document.getElementById("springConstantLabel").textContent = `Spring Constant: ${this.springConstant}N/m`;

    document.getElementById("animationDurationLabel").textContent = `Animation Duration: ${this.animationDuration}ms`;
    document.getElementById("animationIntensityLabel").textContent = `Animation Intensity: ${this.animationIntensity}`;
  }

  setupCanvas() {
    const resizeCanvas = () => {
      const rect = this.canvas.getBoundingClientRect();
      this.canvas.width = rect.width * window.devicePixelRatio;
      this.canvas.height = rect.height * window.devicePixelRatio;
      this.ctx.scale(window.devicePixelRatio, window.devicePixelRatio);
      this.canvas.style.width = rect.width + "px";
      this.canvas.style.height = rect.height + "px";
    };

    window.addEventListener("resize", resizeCanvas);
    resizeCanvas();
  }

  setupEventListeners() {
    // Canvas mouse events
    this.canvas.addEventListener("mousedown", (e) => this.onMouseDown(e));
    this.canvas.addEventListener("mousemove", (e) => this.onMouseMove(e));
    this.canvas.addEventListener("mouseup", (e) => this.onMouseUp(e));
    this.canvas.addEventListener("wheel", (e) => this.onWheel(e));

    // Sidebar controls
    document.getElementById("showLabels").addEventListener("change", (e) => {
      this.showLabels = e.target.checked;
    });

    document.getElementById("showEdgeWeights").addEventListener("change", (e) => {
      this.showEdgeWeights = e.target.checked;
    });

    document.getElementById("repulsionSlider").addEventListener("input", (e) => {
      this.repulsionForce = parseInt(e.target.value);
      this.updateSliderLabels();
    });

    document.getElementById("linkDistanceSlider").addEventListener("input", (e) => {
      this.linkDistance = parseInt(e.target.value);
      this.updateSliderLabels();
    });

    document.getElementById("springConstantSlider").addEventListener("input", (e) => {
      this.springConstant = parseFloat(e.target.value);
      this.updateSliderLabels();
    });

    document.getElementById("resetLayout").addEventListener("click", () => {
      this.resetLayout();
    });

    document.getElementById("centerGraph").addEventListener("click", () => {
      this.centerGraph();
    });

    document.getElementById("addRandomNode").addEventListener("click", () => {
      this.addRandomNode();
    });

    document.getElementById("addRandomEdge").addEventListener("click", () => {
      this.addRandomEdge();
    });

    document.getElementById("clearGraph").addEventListener("click", () => {
      this.clearGraph();
    });

    // Animation controls
    document.getElementById("animationDurationSlider").addEventListener("input", (e) => {
      this.animationDuration = parseInt(e.target.value);
      this.updateSliderLabels();
    });

    document.getElementById("animationIntensitySlider").addEventListener("input", (e) => {
      this.animationIntensity = parseInt(e.target.value);
      this.updateSliderLabels();
    });

    document.getElementById("triggerFlow").addEventListener("click", () => {
      this.triggerRandomFlow();
    });

    document.getElementById("autoFlow").addEventListener("change", (e) => {
      this.autoFlow = e.target.checked;
    });
  }

  generateSampleData() {
    // Create sample users (nodes)
    const users = [
      { id: 1, name: "Alice", color: "#e74c3c" },
      { id: 2, name: "Bob", color: "#3498db" },
      { id: 3, name: "Charlie", color: "#2ecc71" },
      { id: 4, name: "Diana", color: "#f39c12" },
      { id: 5, name: "Eve", color: "#9b59b6" },
      { id: 6, name: "Frank", color: "#1abc9c" },
    ];

    users.forEach((user) => {
      this.addNode(user.name, user.color);
    });

    // Create sample message connections (edges)
    const connections = [
      { from: 0, to: 1, weight: 15 },
      { from: 1, to: 2, weight: 8 },
      { from: 2, to: 3, weight: 12 },
      { from: 0, to: 3, weight: 5 },
      { from: 1, to: 4, weight: 20 },
      { from: 4, to: 5, weight: 7 },
      { from: 3, to: 5, weight: 10 },
      { from: 0, to: 5, weight: 3 },
    ];

    connections.forEach((conn) => {
      this.addEdge(conn.from, conn.to, conn.weight);
    });

    this.updateStats();
  }

  addNode(name, color = null) {
    const canvasRect = this.canvas.getBoundingClientRect();
    const node = {
      id: this.nodes.length,
      name: name,
      x: Math.random() * (canvasRect.width - 100) + 50,
      y: Math.random() * (canvasRect.height - 100) + 50,
      vx: 0,
      vy: 0,
      radius: 20,
      color: color || this.getRandomColor(),
      fixed: false,
    };
    this.nodes.push(node);
    this.updateStats();
  }

  addEdge(fromIndex, toIndex, weight = 1) {
    if (fromIndex < this.nodes.length && toIndex < this.nodes.length && fromIndex !== toIndex) {
      const edge = {
        from: fromIndex,
        to: toIndex,
        weight: weight,
      };
      this.edges.push(edge);
      this.updateStats();
    }
  }

  constainsEdge(src, dest) {
    for (const edge of this.edges) {
      if ((edge.from == src && edge.to == dest) || (edge.from == dest && edge.to == src)) {
        return true;
      }
    }

    return false;
  }

  getRandomColor() {
    const colors = ["#e74c3c", "#3498db", "#2ecc71", "#f39c12", "#9b59b6", "#1abc9c", "#e67e22", "#34495e"];
    return colors[Math.floor(Math.random() * colors.length)];
  }

  getRandomName() {
    const names = ["Alex", "Sam", "Jordan", "Taylor", "Casey", "Morgan", "Riley", "Avery", "Quinn", "Sage"];
    return names[Math.floor(Math.random() * names.length)] + Math.floor(Math.random() * 1000);
  }

  addRandomNode() {
    this.addNode(this.getRandomName());
  }

  addRandomEdge() {
    if (this.nodes.length < 2) return;

    while (true) {
      const src = Math.floor(Math.random() * this.nodes.length);
      const dest = Math.floor(Math.random() * this.nodes.length);

      if (src != dest && !this.constainsEdge(src, dest)) {
        const weight = Math.floor(Math.random() * 20) + 1;
        this.addEdge(src, dest, weight);
        return;
      }
    }
    // while (to === from) {
    //   to = Math.floor(Math.random() * this.nodes.length);
    // }
  }

  clearGraph() {
    this.nodes = [];
    this.edges = [];
    this.updateStats();
  }

  resetLayout() {
    const canvasRect = this.canvas.getBoundingClientRect();
    this.nodes.forEach((node) => {
      node.x = Math.random() * (canvasRect.width - 100) + 50;
      node.y = Math.random() * (canvasRect.height - 100) + 50;
      node.vx = 0;
      node.vy = 0;
    });
  }

  centerGraph() {
    if (this.nodes.length === 0) return;

    const canvasRect = this.canvas.getBoundingClientRect();
    const centerX = canvasRect.width / 2;
    const centerY = canvasRect.height / 2;

    // Calculate current center of nodes
    let avgX = 0,
      avgY = 0;
    this.nodes.forEach((node) => {
      avgX += node.x;
      avgY += node.y;
    });
    avgX /= this.nodes.length;
    avgY /= this.nodes.length;

    // Move all nodes to center the graph
    const deltaX = centerX - avgX;
    const deltaY = centerY - avgY;

    this.nodes.forEach((node) => {
      node.x += deltaX;
      node.y += deltaY;
    });
  }

  updateStats() {
    document.getElementById("nodeCount").textContent = `Nodes: ${this.nodes.length}`;
    document.getElementById("edgeCount").textContent = `Edges: ${this.edges.length}`;
  }

  // Message flow animation methods
  triggerRandomFlow() {
    if (this.edges.length === 0) return;

    const randomEdge = this.edges[Math.floor(Math.random() * this.edges.length)];
    this.createMessageFlow(randomEdge.from, randomEdge.to);
  }

  createMessageFlow(fromNodeIndex, toNodeIndex) {
    const flow = {
      from: fromNodeIndex,
      to: toNodeIndex,
      progress: 0,
      startTime: Date.now(),
      duration: this.animationDuration, // Use duration directly from slider
      color: this.getFlowColor(),
    };
    this.messageFlows.push(flow);
  }

  getFlowColor() {
    const colors = ["#ff6b6b", "#4ecdc4", "#45b7d1", "#96ceb4", "#feca57", "#ff9ff3", "#54a0ff"];
    return colors[Math.floor(Math.random() * colors.length)];
  }

  updateMessageFlows() {
    const currentTime = Date.now();

    // Auto-trigger flows if enabled
    if (this.autoFlow && currentTime - this.lastAutoFlowTime > 200) {
      this.triggerRandomFlow();
      this.lastAutoFlowTime = currentTime;
    }

    // Update existing flows
    this.messageFlows = this.messageFlows.filter((flow) => {
      const elapsed = currentTime - flow.startTime;
      if (elapsed > flow.duration) {
        return false;
      }

      flow.progress = easeIn(elapsed / flow.duration);
      return true;
    });
  }

  onMouseDown(e) {
    const rect = this.canvas.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    // Check if clicking on a node
    for (let node of this.nodes) {
      const dx = mouseX - node.x;
      const dy = mouseY - node.y;
      const distance = Math.sqrt(dx * dx + dy * dy);

      if (distance <= node.radius) {
        this.isDragging = true;
        this.dragNode = node;
        node.fixed = true;
        this.lastMouseX = mouseX;
        this.lastMouseY = mouseY;
        return;
      }
    }

    // If not clicking on a node, prepare for canvas panning
    this.isDragging = true;
    this.lastMouseX = mouseX;
    this.lastMouseY = mouseY;
  }

  onMouseMove(e) {
    if (!this.isDragging) return;

    const rect = this.canvas.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    if (this.dragNode) {
      // Drag node
      this.dragNode.x = mouseX;
      this.dragNode.y = mouseY;
      this.dragNode.vx = 0;
      this.dragNode.vy = 0;
    } else {
      // Pan canvas
      const deltaX = mouseX - this.lastMouseX;
      const deltaY = mouseY - this.lastMouseY;

      this.nodes.forEach((node) => {
        node.x += deltaX;
        node.y += deltaY;
      });
    }

    this.lastMouseX = mouseX;
    this.lastMouseY = mouseY;
  }

  onMouseUp(e) {
    if (this.dragNode) {
      this.dragNode.fixed = false;
      this.dragNode = null;
    }
    this.isDragging = false;
  }

  onWheel(e) {
    e.preventDefault();
    // Zoom functionality can be added here
  }

  applyForces() {
    // Reset forces
    this.nodes.forEach((node) => {
      if (!node.fixed) {
        node.fx = 0;
        node.fy = 0;
      }
    });

    // Repulsion force between nodes
    for (let i = 0; i < this.nodes.length; i++) {
      for (let j = i + 1; j < this.nodes.length; j++) {
        const nodeA = this.nodes[i];
        const nodeB = this.nodes[j];

        const dx = nodeB.x - nodeA.x;
        const dy = nodeB.y - nodeA.y;
        const distance = Math.sqrt(dx * dx + dy * dy);

        if (distance > 0) {
          const force = this.repulsionForce / (distance * distance);
          const fx = (dx / distance) * force;
          const fy = (dy / distance) * force;

          if (!nodeA.fixed) {
            nodeA.fx = (nodeA.fx || 0) - fx;
            nodeA.fy = (nodeA.fy || 0) - fy;
          }
          if (!nodeB.fixed) {
            nodeB.fx = (nodeB.fx || 0) + fx;
            nodeB.fy = (nodeB.fy || 0) + fy;
          }
        }
      }
    }

    // Attraction force along edges
    this.edges.forEach((edge) => {
      const nodeA = this.nodes[edge.from];
      const nodeB = this.nodes[edge.to];

      const dx = nodeB.x - nodeA.x;
      const dy = nodeB.y - nodeA.y;
      const distance = Math.sqrt(dx * dx + dy * dy);

      const targetDistance = this.linkDistance;
      const force = (distance - targetDistance) * this.springConstant;

      if (distance > 0) {
        const fx = (dx / distance) * force;
        const fy = (dy / distance) * force;

        if (!nodeA.fixed) {
          nodeA.fx = (nodeA.fx || 0) + fx;
          nodeA.fy = (nodeA.fy || 0) + fy;
        }
        if (!nodeB.fixed) {
          nodeB.fx = (nodeB.fx || 0) - fx;
          nodeB.fy = (nodeB.fy || 0) - fy;
        }
      }
    });

    // Apply forces to velocities and positions
    this.nodes.forEach((node) => {
      if (!node.fixed) {
        node.vx = (node.vx + (node.fx || 0)) * this.damping;
        node.vy = (node.vy + (node.fy || 0)) * this.damping;

        node.x += node.vx;
        node.y += node.vy;

        // Keep nodes within canvas bounds
        const rect = this.canvas.getBoundingClientRect();
        node.x = Math.max(node.radius, Math.min(rect.width - node.radius, node.x));
        node.y = Math.max(node.radius, Math.min(rect.height - node.radius, node.y));
      }
    });
  }

  draw() {
    const rect = this.canvas.getBoundingClientRect();
    this.ctx.clearRect(0, 0, rect.width, rect.height);

    // Draw edges
    this.ctx.strokeStyle = "rgba(255, 255, 255, 0.6)";
    this.ctx.lineWidth = 1;

    this.edges.forEach((edge, edgeIndex) => {
      const nodeA = this.nodes[edge.from];
      const nodeB = this.nodes[edge.to];

      // Adjust line width based on edge weight
      this.ctx.lineWidth = Math.max(1, edge.weight / 10);

      this.ctx.beginPath();
      this.ctx.moveTo(nodeA.x, nodeA.y);
      this.ctx.lineTo(nodeB.x, nodeB.y);
      this.ctx.stroke();

      // Draw edge weight if enabled
      if (this.showEdgeWeights) {
        const midX = (nodeA.x + nodeB.x) / 2;
        const midY = (nodeA.y + nodeB.y) / 2;

        this.ctx.fillStyle = "rgba(255, 255, 255, 0.8)";
        this.ctx.font = "12px Arial";
        this.ctx.textAlign = "center";
        this.ctx.fillText(edge.weight.toString(), midX, midY);
      }
    });

    // Draw message flows
    this.messageFlows.forEach((flow) => {
      const nodeA = this.nodes[flow.from];
      const nodeB = this.nodes[flow.to];

      if (!nodeA || !nodeB) return;

      const dx = nodeB.x - nodeA.x;
      const dy = nodeB.y - nodeA.y;

      const distance = Math.sqrt(dx * dx + dy * dy);

      // const perpendicular

      // Calculate current position along the edge
      const currentX = nodeA.x + dx * flow.progress;
      const currentY = nodeA.y + dy * flow.progress;

      const startX = currentX - ((dx / distance) * flowLength) / 2;
      const startY = currentY - ((dy / distance) * flowLength) / 2;

      const endX = currentX + ((dx / distance) * flowLength) / 2;
      const endY = currentY + ((dy / distance) * flowLength) / 2;

      const triangleBase = 3;

      const vertex1X = endX - ((dy / distance) * triangleBase) / 2;
      const vertex1Y = endY + ((dx / distance) * triangleBase) / 2;
      const vertex2X = endX + ((dy / distance) * triangleBase) / 2;
      const vertex2Y = endY - ((dx / distance) * triangleBase) / 2;

      // Calculate intensity-based opacity and size
      // const opacity = this.animationIntensity / 100;
      // const pulseSize = 8 + (this.animationIntensity / 100) * 12;
      // const pulseSize = this.animationIntensity;

      // Create gradient effect
      // const gradient = this.ctx.createRadialGradient(currentX, currentY, 0, currentX, currentY, pulseSize);

      // Parse color and add opacity
      // const color = flow.color;
      const color = "white";
      // const gradient = this.ctx.createLinearGradient(startX, startY, endX, endY);
      // gradient.addColorStop(0, "white");
      // gradient.addColorStop(1, "transparent");

      // gradient.addColorStop(0, color.replace("rgb", "rgba").replace(")", `, ${opacity})`));
      // gradient.addColorStop(0.5, color.replace("rgb", "rgba").replace(")", `, ${opacity * 0.6})`));
      // gradient.addColorStop(1, color.replace("rgb", "rgba").replace(")", `, 0)`));

      // Draw the animated flow
      this.ctx.lineWidth = 1;
      this.ctx.strokeStyle = color;
      this.ctx.fillStyle = color;
      this.ctx.beginPath();
      this.ctx.moveTo(startX, startY);
      // this.ctx.lineTo(endX, endY);
      this.ctx.lineTo(vertex1X, vertex1Y);
      this.ctx.lineTo(vertex2X, vertex2Y);
      this.ctx.lineTo(startX, startY);
      this.ctx.stroke();
      this.ctx.fill();
      //
      //
      // this.ctx.fillStyle = gradient;
      // this.ctx.beginPath();
      // this.ctx.arc(currentX, currentY, pulseSize, 0, 2 * Math.PI);
      // this.ctx.fill();

      // Add a bright center
      // this.ctx.fillStyle = color;
      // this.ctx.beginPath();
      // this.ctx.arc(currentX, currentY, pulseSize * 0.3, 0, 2 * Math.PI);
      // this.ctx.fill();
    });

    // Draw nodes
    this.nodes.forEach((node) => {
      // Draw node circle
      this.ctx.fillStyle = node.color;
      this.ctx.beginPath();
      this.ctx.arc(node.x, node.y, node.radius, 0, 2 * Math.PI);
      this.ctx.fill();

      // Draw node border
      this.ctx.strokeStyle = "rgba(255, 255, 255, 0.8)";
      this.ctx.lineWidth = 2;
      this.ctx.stroke();

      // Draw node label if enabled
      if (this.showLabels) {
        this.ctx.fillStyle = "white";
        this.ctx.font = "14px Arial";
        this.ctx.textAlign = "center";
        this.ctx.fillText(node.name, node.x, node.y + 5);
      }
    });
  }

  startAnimation() {
    const animate = () => {
      this.updateMessageFlows();
      this.applyForces();
      this.draw();
      requestAnimationFrame(animate);
    };
    animate();
  }
}

// Initialize the visualization when the page loads
document.addEventListener("DOMContentLoaded", () => {
  new MessageGraphVisualization();
});
