// The single script file for the entire project
//

const flowLength = 30; // px
const flowsPerParticlePerSecond = 2;

// webgl textures which hold vertex locations cannot be resized easily,
// so just allocated a bunch of space and hope we dont need it all
const MAX_PARTICLE_COUNT = 1024;

const MAX_EDGE_COUNT = 4 * MAX_PARTICLE_COUNT;

// the floats needed for a webgl vertex
const FLOATS_PER_VERTEX = 4;

// Canvas Manager Class - handles switching between GPU and CPU backends
class CanvasManager {
  constructor(containerId) {
    this.container = document.getElementById(containerId);
    this.cpuCanvas = document.getElementById("cpu-canvas");
    this.gpuCanvas = document.getElementById("gpu-canvas");

    this.cpuCtx = null;
    this.gpuCtx = null;

    this.currentBackend = "cpu";
    this.isInitialized = false;

    this.init();
  }

  init() {
    // Initialize CPU context (2D)
    this.cpuCtx = this.cpuCanvas.getContext("2d");

    // Try to initialize GPU context (WebGL2)
    try {
      this.gpuCtx = this.gpuCanvas.getContext("webgl2");
      if (!this.gpuCtx) {
        console.warn("WebGL2 not supported, GPU backend disabled");
      }
    } catch (error) {
      console.warn("Failed to get WebGL2 context:", error);
    }

    // Set up canvas sizing
    this.setupCanvasSizing();

    // Initially show CPU canvas
    this.switchToBackend("cpu");

    this.isInitialized = true;
  }

  setupCanvasSizing() {
    const resizeCanvases = () => {
      const rect = this.container.getBoundingClientRect();
      const dpr = window.devicePixelRatio || 1;

      [this.cpuCanvas, this.gpuCanvas].forEach((canvas) => {
        canvas.width = rect.width * dpr;
        canvas.height = rect.height * dpr;
        canvas.style.width = rect.width + "px";
        canvas.style.height = rect.height + "px";
      });

      // Scale CPU context for high DPI
      this.cpuCtx.scale(dpr, dpr);
    };

    window.addEventListener("resize", resizeCanvases);
    resizeCanvases();
  }

  switchToBackend(backend) {
    if (!this.isInitialized) return false;

    if (backend === "gpu" && !this.gpuCtx) {
      console.warn("GPU backend not available, staying on CPU");
      return false;
    }

    this.currentBackend = backend;

    if (backend === "cpu") {
      this.cpuCanvas.classList.remove("hidden");
      this.gpuCanvas.classList.add("hidden");
    } else {
      this.cpuCanvas.classList.add("hidden");
      this.gpuCanvas.classList.remove("hidden");
    }

    return true;
  }

  getActiveCanvas() {
    return this.currentBackend === "cpu" ? this.cpuCanvas : this.gpuCanvas;
  }

  getActiveContext() {
    return this.currentBackend === "cpu" ? this.cpuCtx : this.gpuCtx;
  }

  getCurrentBackend() {
    return this.currentBackend;
  }

  isGPUAvailable() {
    return this.gpuCtx !== null;
  }

  // Forward canvas events to the active canvas
  addEventListener(event, handler) {
    this.cpuCanvas.addEventListener(event, handler);
    this.gpuCanvas.addEventListener(event, handler);
  }

  removeEventListener(event, handler) {
    this.cpuCanvas.removeEventListener(event, handler);
    this.gpuCanvas.removeEventListener(event, handler);
  }

  getBoundingClientRect() {
    return this.getActiveCanvas().getBoundingClientRect();
  }
}

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

// compiles a webgl shader given a source string
function compileShader(gl, type, source) {
  const shader = gl.createShader(type);
  gl.shaderSource(shader, source);
  gl.compileShader(shader);
  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    console.error(gl.getShaderInfoLog(shader));
    gl.deleteShader(shader);
    return null;
  }
  return shader;
}

// helper function to compile a program
// (theres a wee bit of boilerplate in WebGL)
function createProgram(gl, vertSrc, fragSrc) {
  const vert = compileShader(gl, gl.VERTEX_SHADER, vertSrc);
  const frag = compileShader(gl, gl.FRAGMENT_SHADER, fragSrc);
  const program = gl.createProgram();
  gl.attachShader(program, vert);
  gl.attachShader(program, frag);
  gl.linkProgram(program);
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    console.error(gl.getProgramInfoLog(program));
    return null;
  }
  return program;
}

// loads a shader source
async function loadShader(path) {
  const res = await fetch(path);
  return res.text();
}

// GPU Physics Simulation Class
class GPUPhysicsSimulation {
  constructor(canvas, gl) {
    this.canvas = canvas;
    this.gl = gl;
    this.program = null;
    this.buffers = {};
    this.textures = {};
    this.uniforms = {};
    this.attributes = {};
    this.framebuffers = {};
    this.tf = gl.createTransformFeedback();

    // Simulation state
    this.nodeCount = 0;
    this.edgeCount = 0;

    this.nodes = [];
    this.edges = [];

    this.initialized = false;

    // Pending async readback from previous frame
    this.pendingReadback = null; // { sync, nodeCount, stagingIdx }
  }

  async init() {
    console.log("initializing...");
    // Get WebGL2 context
    // this.gl = this.canvas.getContext("webgl2");
    // if (!this.gl) {
    //   throw new Error("WebGL2 not supported");
    // }

    // Load and compile shaders
    const vertexShaderSource = await loadShader("shaders/physics-vertex.glsl");
    const fragmentShaderSource = await loadShader("shaders/physics-fragment.glsl");

    this.program = this.createPhysicsProgram(vertexShaderSource, fragmentShaderSource);
    if (!this.program) {
      throw new Error("Failed to create physics shader program");
    }

    // Get uniform and attribute locations
    this.getLocations();

    // Initialize buffers and textures
    console.log("bufetting...");
    this.initBuffers();

    this.initialized = true;
  }

  createPhysicsProgram(vertSource, fragSource) {
    const program = createProgram(this.gl, vertSource, fragSource);

    if (!program) return null;

    // Set up transform feedback
    const varyings = ["v_newPosition", "v_newVelocity"];
    this.gl.transformFeedbackVaryings(program, varyings, this.gl.INTERLEAVED_ATTRIBS);
    this.gl.linkProgram(program);

    if (!this.gl.getProgramParameter(program, this.gl.LINK_STATUS)) {
      console.error("Transform feedback program link failed:", this.gl.getProgramInfoLog(program));
      return null;
    }

    return program;
  }

  getLocations() {
    this.gl.useProgram(this.program);

    // Attribute locations
    this.attributes = {
      position: this.gl.getAttribLocation(this.program, "a_position"),
      velocity: this.gl.getAttribLocation(this.program, "a_velocity"),
      mass: this.gl.getAttribLocation(this.program, "a_mass"),
      radius: this.gl.getAttribLocation(this.program, "a_radius"),
      fixed: this.gl.getAttribLocation(this.program, "a_fixed"),
    };

    // Uniform locations
    this.uniforms = {
      repulsionForce: this.gl.getUniformLocation(this.program, "u_repulsionForce"),
      springConstant: this.gl.getUniformLocation(this.program, "u_springConstant"),
      linkDistance: this.gl.getUniformLocation(this.program, "u_linkDistance"),
      damping: this.gl.getUniformLocation(this.program, "u_damping"),
      deltaTime: this.gl.getUniformLocation(this.program, "u_deltaTime"),
      canvasSize: this.gl.getUniformLocation(this.program, "u_canvasSize"),
      nodePositions: this.gl.getUniformLocation(this.program, "u_nodePositions"),
      nodeData: this.gl.getUniformLocation(this.program, "u_nodeData"),
      nodeCount: this.gl.getUniformLocation(this.program, "u_nodeCount"),
      edgeData: this.gl.getUniformLocation(this.program, "u_edgeData"),
      edgeCount: this.gl.getUniformLocation(this.program, "u_edgeCount"),
    };
  }

  initBuffers() {
    const gl = this.gl;

    // Create vertex array objects for double buffering
    this.vaos = {
      read: gl.createVertexArray(),
      write: gl.createVertexArray(),
    };

    // Create transform feedback objects
    this.transformFeedbacks = [gl.createBuffer(), gl.createBuffer()];

    const tfSize = MAX_PARTICLE_COUNT * 2 * 4;
    gl.bindBuffer(gl.ARRAY_BUFFER, this.transformFeedbacks[0]);
    gl.bufferData(gl.ARRAY_BUFFER, tfSize, gl.DYNAMIC_DRAW);

    gl.bindBuffer(gl.ARRAY_BUFFER, this.transformFeedbacks[1]);
    gl.bufferData(gl.ARRAY_BUFFER, tfSize, gl.DYNAMIC_DRAW);

    // Initialize buffers object
    // this.buffers = {};
    // this.textures = {};

    // Initialize with empty buffers - will be resized when nodes are added
    this.resizeBuffers(0, 0);
  }

  resizeBuffers(nodeCount, edgeCount) {
    console.log("resizing...");
    const gl = this.gl;

    this.nodeCount = nodeCount;
    this.edgeCount = edgeCount;

    if (nodeCount === 0) return;

    // Calculate texture dimensions for node data
    const nodeTexWidth = Math.ceil(Math.sqrt(nodeCount));
    const nodeTexHeight = Math.ceil(nodeCount / nodeTexWidth);

    // Calculate texture dimensions for edge data
    const edgeTexWidth = Math.max(1, Math.ceil(Math.sqrt(edgeCount)));
    const edgeTexHeight = Math.max(1, Math.ceil(edgeCount / edgeTexWidth));

    // Create or update node position buffers (double buffered)
    if (this.buffers.nodePositions) {
      gl.deleteBuffer(this.buffers.nodePositions[0]);
      gl.deleteBuffer(this.buffers.nodePositions[1]);
    }
    this.buffers.nodePositions = [gl.createBuffer(), gl.createBuffer()];

    // Create or update node velocity buffers (double buffered)
    if (this.buffers.nodeVelocities) {
      gl.deleteBuffer(this.buffers.nodeVelocities[0]);
      gl.deleteBuffer(this.buffers.nodeVelocities[1]);
    }
    this.buffers.nodeVelocities = [gl.createBuffer(), gl.createBuffer()];

    // Create node attribute buffers
    if (this.buffers.nodeAttributes) {
      gl.deleteBuffer(this.buffers.nodeAttributes);
    }
    this.buffers.nodeAttributes = gl.createBuffer();

    // Initialize buffers with proper size
    const positionBufferSize = nodeCount * FLOATS_PER_VERTEX * 4; // 4 floats per vertex * 4 bytes per float
    const velocityBufferSize = nodeCount * FLOATS_PER_VERTEX * 4; // 4 floats per vertex * 4 bytes per float
    const attributeBufferSize = nodeCount * FLOATS_PER_VERTEX * 4; // 4 floats per vertex * 4 bytes per float

    // Allocate position buffers
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.nodePositions[0]);
    gl.bufferData(gl.ARRAY_BUFFER, positionBufferSize, gl.DYNAMIC_DRAW);

    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.nodePositions[1]);
    gl.bufferData(gl.ARRAY_BUFFER, positionBufferSize, gl.DYNAMIC_DRAW);

    // Allocate velocity buffers
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.nodeVelocities[0]);
    gl.bufferData(gl.ARRAY_BUFFER, velocityBufferSize, gl.DYNAMIC_DRAW);

    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.nodeVelocities[1]);
    gl.bufferData(gl.ARRAY_BUFFER, velocityBufferSize, gl.DYNAMIC_DRAW);

    // Allocate attribute buffer
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.nodeAttributes);
    gl.bufferData(gl.ARRAY_BUFFER, attributeBufferSize, gl.STATIC_DRAW);

    // Create or update textures for node data
    if (this.textures.nodePositions) {
      gl.deleteTexture(this.textures.nodePositions);
    }
    this.textures.nodePositions = this.createDataTexture(nodeTexWidth, nodeTexHeight);

    if (this.textures.nodeData) {
      gl.deleteTexture(this.textures.nodeData);
    }
    this.textures.nodeData = this.createDataTexture(nodeTexWidth, nodeTexHeight);

    // Create or update edge data texture
    if (this.textures.edgeData) {
      gl.deleteTexture(this.textures.edgeData);
    }
    this.textures.edgeData = this.createDataTexture(edgeTexWidth, edgeTexHeight);

    // Create staging buffers for async readback (STREAM_READ avoids pipeline stalls)
    if (this.buffers.stagingOutput) {
      gl.deleteBuffer(this.buffers.stagingOutput[0]);
      gl.deleteBuffer(this.buffers.stagingOutput[1]);
    }
    this.buffers.stagingOutput = [gl.createBuffer(), gl.createBuffer()];
    const stagingSize = nodeCount * 4 * 4; // 4 floats per node (pos xy + vel xy) * 4 bytes
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.stagingOutput[0]);
    gl.bufferData(gl.ARRAY_BUFFER, stagingSize, gl.STREAM_READ);
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.stagingOutput[1]);
    gl.bufferData(gl.ARRAY_BUFFER, stagingSize, gl.STREAM_READ);
    gl.bindBuffer(gl.ARRAY_BUFFER, null);

    // Invalidate any pending readback since buffer layout changed
    if (this.pendingReadback) {
      gl.deleteSync(this.pendingReadback.sync);
      this.pendingReadback = null;
    }

    this.currentBuffer = 0; // For double buffering
  }

  createDataTexture(width, height) {
    const gl = this.gl;
    const texture = gl.createTexture();

    gl.bindTexture(gl.TEXTURE_2D, texture);
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, width, height, 0, gl.RGBA, gl.FLOAT, null);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);

    return texture;
  }

  updateData(nodes, edges, particleMass) {
    if (!this.initialized || nodes.length === 0) return;

    // Resize buffers if needed
    if (nodes.length !== this.nodeCount || edges.length !== this.edgeCount) {
      this.resizeBuffers(nodes.length, edges.length);
    }

    const gl = this.gl;

    // Update node position and velocity data
    const positionData = new Float32Array(nodes.length * 4);
    const velocityData = new Float32Array(nodes.length * 4);
    const attributeData = new Float32Array(nodes.length * 4);

    for (let i = 0; i < nodes.length; i++) {
      const node = nodes[i];
      const base = i * 4;

      // Position data (x, y, 0, 0)
      positionData[base] = node.x;
      positionData[base + 1] = node.y;
      positionData[base + 2] = 0.0;
      positionData[base + 3] = 0.0;

      // Velocity data (vx, vy, 0, 0)
      velocityData[base] = node.vx || 0.0;
      velocityData[base + 1] = node.vy || 0.0;
      velocityData[base + 2] = 0.0;
      velocityData[base + 3] = 0.0;

      // Attribute data (mass, radius, fixed, 0)
      attributeData[base] = particleMass; // mass
      attributeData[base + 1] = node.radius;
      attributeData[base + 2] = node.fixed ? 1.0 : 0.0;
      attributeData[base + 3] = 0.0;
    }

    // Update position buffers
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.nodePositions[this.currentBuffer]);
    gl.bufferData(gl.ARRAY_BUFFER, positionData, gl.DYNAMIC_DRAW);

    // Update velocity buffers
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.nodeVelocities[this.currentBuffer]);
    gl.bufferData(gl.ARRAY_BUFFER, velocityData, gl.DYNAMIC_DRAW);

    // Update attribute buffer
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.nodeAttributes);
    gl.bufferData(gl.ARRAY_BUFFER, attributeData, gl.DYNAMIC_DRAW);

    // Update node position texture
    this.updateNodeTextures(nodes, particleMass);

    // Update edge data texture
    this.updateEdgeTexture(edges);
  }

  updateNodeTextures(nodes, particleMass = 10.0) {
    const gl = this.gl;
    const nodeTexWidth = Math.ceil(Math.sqrt(nodes.length));
    const nodeTexHeight = Math.ceil(nodes.length / nodeTexWidth);

    // Prepare position texture data
    const posTexData = new Float32Array(nodeTexWidth * nodeTexHeight * 4);
    const dataTexData = new Float32Array(nodeTexWidth * nodeTexHeight * 4);

    for (let i = 0; i < nodes.length; i++) {
      const node = nodes[i];
      const base = i * 4;

      posTexData[base] = node.x;
      posTexData[base + 1] = node.y;
      posTexData[base + 2] = 0.0;
      posTexData[base + 3] = 0.0;

      dataTexData[base] = particleMass; // mass
      dataTexData[base + 1] = node.radius;
      dataTexData[base + 2] = node.fixed ? 1.0 : 0.0;
      dataTexData[base + 3] = 0.0;
    }

    // Update position texture
    gl.bindTexture(gl.TEXTURE_2D, this.textures.nodePositions);
    gl.texSubImage2D(gl.TEXTURE_2D, 0, 0, 0, nodeTexWidth, nodeTexHeight, gl.RGBA, gl.FLOAT, posTexData);

    // Update data texture
    gl.bindTexture(gl.TEXTURE_2D, this.textures.nodeData);
    gl.texSubImage2D(gl.TEXTURE_2D, 0, 0, 0, nodeTexWidth, nodeTexHeight, gl.RGBA, gl.FLOAT, dataTexData);
  }

  updateEdgeTexture(edges) {
    if (edges.length === 0) return;

    const gl = this.gl;
    const edgeTexWidth = Math.ceil(Math.sqrt(edges.length));
    const edgeTexHeight = Math.ceil(edges.length / edgeTexWidth);

    // Prepare edge texture data
    const edgeTexData = new Float32Array(edgeTexWidth * edgeTexHeight * 4);

    for (let i = 0; i < edges.length; i++) {
      const edge = edges[i];
      const base = i * 4;

      edgeTexData[base] = edge.from;
      edgeTexData[base + 1] = edge.to;
      edgeTexData[base + 2] = edge.weight;
      edgeTexData[base + 3] = 0.0;
    }

    // Update edge texture
    gl.bindTexture(gl.TEXTURE_2D, this.textures.edgeData);
    gl.texSubImage2D(gl.TEXTURE_2D, 0, 0, 0, edgeTexWidth, edgeTexHeight, gl.RGBA, gl.FLOAT, edgeTexData);
  }

  simulate(nodes, edges, params) {
    if (!this.initialized || nodes.length === 0) return;

    // Apply previous frame's readback before overwriting node data
    this.applyPendingReadback(nodes);

    const gl = this.gl;
    const rect = this.canvas.getBoundingClientRect();

    // Update data
    this.updateData(nodes, edges, params.particleMass);

    // Use physics program
    gl.useProgram(this.program);

    // Set uniforms
    gl.uniform1f(this.uniforms.repulsionForce, params.repulsionForce);
    gl.uniform1f(this.uniforms.springConstant, params.springConstant);
    gl.uniform1f(this.uniforms.linkDistance, params.linkDistance);
    gl.uniform1f(this.uniforms.damping, params.damping);
    gl.uniform1f(this.uniforms.deltaTime, params.deltaTime || 1.0);
    gl.uniform2f(this.uniforms.canvasSize, rect.width, rect.height);
    gl.uniform1i(this.uniforms.nodeCount, nodes.length);
    gl.uniform1i(this.uniforms.edgeCount, edges.length);

    // Bind textures
    gl.activeTexture(gl.TEXTURE0);
    gl.bindTexture(gl.TEXTURE_2D, this.textures.nodePositions);
    gl.uniform1i(this.uniforms.nodePositions, 0);

    gl.activeTexture(gl.TEXTURE1);
    gl.bindTexture(gl.TEXTURE_2D, this.textures.nodeData);
    gl.uniform1i(this.uniforms.nodeData, 1);

    gl.activeTexture(gl.TEXTURE2);
    gl.bindTexture(gl.TEXTURE_2D, this.textures.edgeData);
    gl.uniform1i(this.uniforms.edgeData, 2);

    // Get buffer references
    const currentPosBuffer = this.buffers.nodePositions[this.currentBuffer];
    const currentVelBuffer = this.buffers.nodeVelocities[this.currentBuffer];
    const nextPosBuffer = this.buffers.nodePositions[1 - this.currentBuffer];
    const nextVelBuffer = this.buffers.nodeVelocities[1 - this.currentBuffer];

    // Bind VAO for input attributes
    gl.bindVertexArray(this.vaos.read);

    // Set up vertex attributes - Position
    gl.bindBuffer(gl.ARRAY_BUFFER, currentPosBuffer);
    if (this.attributes.position >= 0) {
      gl.enableVertexAttribArray(this.attributes.position);
      gl.vertexAttribPointer(this.attributes.position, 2, gl.FLOAT, false, 16, 0);
    }

    // Set up vertex attributes - Velocity
    gl.bindBuffer(gl.ARRAY_BUFFER, currentVelBuffer);
    if (this.attributes.velocity >= 0) {
      gl.enableVertexAttribArray(this.attributes.velocity);
      gl.vertexAttribPointer(this.attributes.velocity, 2, gl.FLOAT, false, 16, 0);
    }

    // Set up vertex attributes - Node attributes (mass, radius, fixed)
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.nodeAttributes);
    if (this.attributes.mass >= 0) {
      gl.enableVertexAttribArray(this.attributes.mass);
      gl.vertexAttribPointer(this.attributes.mass, 1, gl.FLOAT, false, 16, 0);
    }
    if (this.attributes.radius >= 0) {
      gl.enableVertexAttribArray(this.attributes.radius);
      gl.vertexAttribPointer(this.attributes.radius, 1, gl.FLOAT, false, 16, 4);
    }
    if (this.attributes.fixed >= 0) {
      gl.enableVertexAttribArray(this.attributes.fixed);
      gl.vertexAttribPointer(this.attributes.fixed, 1, gl.FLOAT, false, 16, 8);
    }

    // Unbind VAO to prevent conflicts with transform feedback
    gl.bindVertexArray(null);

    // Set up transform feedback for output
    gl.bindTransformFeedback(gl.TRANSFORM_FEEDBACK, this.tf);

    // Create interleaved buffer for transform feedback output
    // if (!this.buffers.transformFeedbackOutput) {
    //   this.buffers.transformFeedbackOutput = [gl.createBuffer(), gl.createBuffer()];
    //
    //   // Allocate buffers for interleaved output (position + velocity)
    //   const outputBufferSize = nodes.length * 8 * 4; // 8 floats per vertex (4 for position, 4 for velocity) * 4 bytes per float

    // gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.transformFeedbackOutput[0]);
    // gl.bufferData(gl.ARRAY_BUFFER, outputBufferSize, gl.DYNAMIC_DRAW);
    //
    // gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.transformFeedbackOutput[1]);
    // gl.bufferData(gl.ARRAY_BUFFER, outputBufferSize, gl.DYNAMIC_DRAW);
    // }

    // Bind the current output buffer for transform feedback
    const currentOutputBuffer = this.transformFeedbacks[this.currentBuffer];
    gl.bindBufferBase(gl.TRANSFORM_FEEDBACK_BUFFER, 0, currentOutputBuffer);

    // Disable rasterization (we only want transform feedback)
    gl.enable(gl.RASTERIZER_DISCARD);

    // Bind input VAO again for drawing
    gl.bindVertexArray(this.vaos.read);

    // Run the physics simulation
    gl.beginTransformFeedback(gl.POINTS);
    gl.drawArrays(gl.POINTS, 0, nodes.length);
    gl.endTransformFeedback();

    // Re-enable rasterization
    gl.disable(gl.RASTERIZER_DISCARD);

    // Clean up bindings
    gl.bindVertexArray(null);
    gl.bindTransformFeedback(gl.TRANSFORM_FEEDBACK, null);

    // Copy TF output to STREAM_READ staging buffer (GPU→GPU, no stall)
    if (this.buffers.stagingOutput) {
      const stagingIdx = this.currentBuffer;
      const byteSize = nodes.length * 4 * 4; // 4 floats per node (vec2 pos + vec2 vel)
      gl.bindBuffer(gl.COPY_READ_BUFFER, currentOutputBuffer);
      gl.bindBuffer(gl.COPY_WRITE_BUFFER, this.buffers.stagingOutput[stagingIdx]);
      gl.copyBufferSubData(gl.COPY_READ_BUFFER, gl.COPY_WRITE_BUFFER, 0, 0, byteSize);
      gl.bindBuffer(gl.COPY_READ_BUFFER, null);
      gl.bindBuffer(gl.COPY_WRITE_BUFFER, null);

      // Replace any unread pending readback
      if (this.pendingReadback) {
        gl.deleteSync(this.pendingReadback.sync);
      }
      this.pendingReadback = {
        sync: gl.fenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0),
        nodeCount: nodes.length,
        stagingIdx,
      };
    }

    // Swap buffers for next frame
    this.currentBuffer = 1 - this.currentBuffer;
  }

  // Apply the readback that was staged on a previous frame, if it's ready.
  // Called at the top of simulate() so results are one frame behind (no stall).
  applyPendingReadback(nodes) {
    if (!this.pendingReadback) return;

    const gl = this.gl;
    const { sync, nodeCount, stagingIdx } = this.pendingReadback;

    // Non-blocking check — 0 timeout returns immediately
    const status = gl.clientWaitSync(sync, 0, 0);
    if (status === gl.TIMEOUT_EXPIRED || status === gl.WAIT_FAILED) return;

    gl.deleteSync(sync);
    this.pendingReadback = null;

    // TF output is interleaved: [px, py, vx, vy] per node (vec2 + vec2)
    const data = new Float32Array(nodeCount * 4);
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.stagingOutput[stagingIdx]);
    gl.getBufferSubData(gl.ARRAY_BUFFER, 0, data);
    gl.bindBuffer(gl.ARRAY_BUFFER, null);

    const count = Math.min(nodeCount, nodes.length);
    for (let i = 0; i < count; i++) {
      // Don't overwrite fixed (dragged) nodes — their position is owned by the mouse handler
      if (nodes[i].fixed) continue;
      const base = i * 4;
      nodes[i].x = data[base];
      nodes[i].y = data[base + 1];
      nodes[i].vx = data[base + 2];
      nodes[i].vy = data[base + 3];
    }
  }

  discardPendingReadback() {
    if (this.pendingReadback) {
      this.gl.deleteSync(this.pendingReadback.sync);
      this.pendingReadback = null;
    }
  }

  cleanup() {
    if (!this.gl) return;

    const gl = this.gl;

    // Clean up buffers
    Object.values(this.buffers).forEach((buffer) => {
      if (Array.isArray(buffer)) {
        buffer.forEach((b) => gl.deleteBuffer(b));
      } else {
        gl.deleteBuffer(buffer);
      }
    });

    // Clean up textures
    Object.values(this.textures).forEach((texture) => {
      gl.deleteTexture(texture);
    });

    // Clean up VAOs
    Object.values(this.vaos).forEach((vao) => {
      gl.deleteVertexArray(vao);
    });

    // Clean up transform feedbacks
    Object.values(this.transformFeedbacks).forEach((tf) => {
      gl.deleteTransformFeedback(tf);
    });

    // Clean up program
    if (this.program) {
      gl.deleteProgram(this.program);
    }

    // Clean up pending async readback
    if (this.pendingReadback) {
      gl.deleteSync(this.pendingReadback.sync);
      this.pendingReadback = null;
    }
  }
}

class MessageGraphVisualization {
  constructor() {
    // Initialize canvas manager for backend switching
    this.canvasManager = new CanvasManager("canvasContainer");
    this.canvas = this.canvasManager.getActiveCanvas();
    this.ctx = this.canvasManager.getActiveContext();

    // Graph data
    this.nodes = [];
    this.edges = [];

    // Simulation parameters (will be read from DOM)
    this.repulsionForce = null;
    this.linkDistance = null;
    this.springConstant = null;
    this.particleMass = null;
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
    this.lastAutoFlowTime = 0;
    this.autoFlowDelay = 1000; // in milliseconds
    this.messageFlows = [];

    // Interaction state
    this.isDragging = false;
    this.dragNode = null;
    this.lastMouseX = 0;
    this.lastMouseY = 0;

    // GPU physics simulation
    this.useGPUPhysics = false;
    this.gpuPhysics = null;

    // Debug state
    this.debugMode = false;
    this.debugTick = 0;
    this.maxDebugTicks = 10;
    this.debugData = [];

    this.init();
  }

  async init() {
    this.setupCanvas();
    this.readInitialValues();
    this.setupEventListeners();

    // Initialize GPU physics system
    try {
      this.gpuPhysics = new GPUPhysicsSimulation(this.canvas, this.canvasManager.gpuCtx);
      await this.gpuPhysics.init();
      console.log("GPU physics simulation initialized successfully");
    } catch (error) {
      console.warn("Failed to initialize GPU physics, falling back to CPU:", error);
      this.useGPUPhysics = false;
    }

    this.generateSampleData();

    this.startAnimation();
  }

  readInitialValues() {
    // Read simulation parameters from sliders
    this.repulsionForce = parseInt(document.getElementById("repulsionSlider").value);
    this.linkDistance = parseInt(document.getElementById("linkDistanceSlider").value);
    this.springConstant = parseFloat(document.getElementById("springConstantSlider").value);
    this.particleMass = parseFloat(document.getElementById("massSlider").value);

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

    document.getElementById("massLabel").textContent = `Particle Mass: ${this.particleMass}kg`;

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
    // Canvas mouse events using canvas manager
    this.canvasManager.addEventListener("mousedown", (e) => this.onMouseDown(e));
    this.canvasManager.addEventListener("mousemove", (e) => this.onMouseMove(e));
    this.canvasManager.addEventListener("mouseup", (e) => this.onMouseUp(e));
    this.canvasManager.addEventListener("wheel", (e) => this.onWheel(e));

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

    document.getElementById("massSlider").addEventListener("input", (e) => {
      this.particleMass = parseFloat(e.target.value);
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

    // GPU physics toggle
    document.getElementById("useGPUPhysics").addEventListener("change", (e) => {
      this.useGPUPhysics = e.target.checked;

      if (this.useGPUPhysics) {
        if (this.gpuPhysics && this.gpuPhysics.initialized) {
          // Discard any stale readback so the first GPU frame starts from
          // the current CPU node positions rather than old GPU positions.
          this.gpuPhysics.discardPendingReadback();
          console.log("Switched to GPU physics simulation");
        } else {
          console.log("GPU physics not available, staying on CPU");
          this.useGPUPhysics = false;
          e.target.checked = false;
        }
      } else {
        console.log("Switched to CPU physics simulation");
      }
    });

    // Sidebar show/hide functionality
    document.getElementById("showSidebar").addEventListener("click", () => {
      this.showSidebar();
    });

    document.getElementById("closeSidebar").addEventListener("click", () => {
      this.hideSidebar();
    });

    // Debug controls
    document.getElementById("resetToKnownState").addEventListener("click", () => {
      this.resetToKnownState();
    });

    document.getElementById("startDebugMode").addEventListener("click", () => {
      this.startDebugMode();
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

    // this.gpuPhysics.nodes.push(node);

    this.autoFlowDelay = (1 / (this.nodes.length * flowsPerParticlePerSecond)) * 1000;

    this.gpuPhysics.resizeBuffers(this.nodes.length, this.edges.length);

    this.updateStats();
  }

  addEdge(fromIndex, toIndex, weight = 1) {
    if (fromIndex < this.nodes.length && toIndex < this.nodes.length && fromIndex !== toIndex) {
      const edge = {
        from: fromIndex,
        to: toIndex,
        weight: weight,
      };
      const reverseEdge = {
        from: toIndex,
        to: fromIndex,
        weight: weight,
      };
      this.edges.push(edge);
      this.edges.push(reverseEdge);

      // this.gpuPhysics.edges.push(edge);
      // this.gpuPhysics.edges.push(reverseEdge);
      this.gpuPhysics.resizeBuffers(this.nodes.length, this.edges.length);

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
    const connectingNode = this.edges[Math.floor(Math.random() * (this.edges.length - 1))].from;

    this.addEdge(connectingNode, this.nodes.length - 1);
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
    if (this.autoFlow && currentTime - this.lastAutoFlowTime > this.autoFlowDelay) {
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
    const rect = this.canvasManager.getBoundingClientRect();
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

    const rect = this.canvasManager.getBoundingClientRect();
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
    if (this.useGPUPhysics && this.gpuPhysics && this.gpuPhysics.initialized) {
      // Use GPU physics simulation
      const params = {
        repulsionForce: this.repulsionForce,
        springConstant: this.springConstant,
        linkDistance: this.linkDistance,
        damping: this.damping,
        deltaTime: 1.0,
        particleMass: this.particleMass,
      };

      this.gpuPhysics.simulate(this.nodes, this.edges, params);
    } else {
      // Use CPU physics simulation (original implementation)
      this.applyCPUForces();
    }
  }

  applyCPUForces() {
    // Physics constants to match GPU implementation
    const deltaTime = 1.0;
    const nodeMass = this.particleMass; // Use configurable mass

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

        if (distance > 0.1) {
          // Match GPU minimum distance threshold
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

      if (distance > 0.1) {
        // Match GPU minimum distance threshold
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

    // Apply forces using proper physics integration (match GPU)
    this.nodes.forEach((node) => {
      if (!node.fixed) {
        // Convert force to acceleration (F = ma, so a = F/m)
        const ax = (node.fx || 0) / nodeMass;
        const ay = (node.fy || 0) / nodeMass;

        // Update velocity with acceleration and apply damping
        node.vx = (node.vx + ax * deltaTime) * this.damping;
        node.vy = (node.vy + ay * deltaTime) * this.damping;

        // Update position with velocity
        node.x += node.vx * deltaTime;
        node.y += node.vy * deltaTime;

        // Keep nodes within canvas bounds
        const rect = this.canvas.getBoundingClientRect();

        // also stop particles if they hit the wall
        if (node.x < node.radius || node.x > rect.width - node.radius) {
          node.vx = 0;
        }
        if (node.y < node.radius || node.y > rect.height - node.radius) {
          node.vy = 0;
        }

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

  showSidebar() {
    const sidebar = document.getElementById("sidebar");
    sidebar.classList.add("active");
  }

  hideSidebar() {
    const sidebar = document.getElementById("sidebar");
    sidebar.classList.remove("active");
  }

  // Debug methods
  resetToKnownState() {
    console.log("Resetting to known state...");

    // Clear existing data
    this.nodes = [];
    this.edges = [];

    // Create 4 nodes in known positions with zero velocity
    const knownNodes = [
      { name: "Node0", x: 500, y: 300, vx: 0, vy: 0 },
      { name: "Node1", x: 400, y: 400, vx: 0, vy: 0 },
      { name: "Node2", x: 600, y: 600, vx: 0, vy: 0 },
      { name: "Node3", x: 500, y: 500, vx: 0, vy: 0 },
    ];

    knownNodes.forEach((nodeData) => {
      const node = {
        id: this.nodes.length,
        name: nodeData.name,
        x: nodeData.x,
        y: nodeData.y,
        vx: nodeData.vx,
        vy: nodeData.vy,
        radius: 20,
        color: "#3498db",
        fixed: false,
      };
      this.nodes.push(node);
    });

    // Add some known edges
    this.addEdge(0, 1, 10); // Node0 <-> Node1
    this.addEdge(1, 2, 15); // Node1 <-> Node2
    this.addEdge(2, 3, 8); // Node2 <-> Node3
    this.addEdge(1, 3, 8); // Node2 <-> Node3

    this.gpuPhysics.discardPendingReadback();

    // Reset debug state
    this.debugTick = 0;
    this.debugData = [];

    console.log("Reset complete. Nodes:", this.nodes.length, "Edges:", this.edges.length);
    this.updateStats();
  }

  startDebugMode() {
    console.log("Starting debug mode...");
    this.debugMode = true;
    this.debugTick = 0;
    this.debugData = [];

    const backend = this.useGPUPhysics ? "GPU" : "CPU";
    console.log(`=== DEBUG MODE STARTED (${backend} Backend) ===`);
    console.log("Simulation Parameters:");
    console.log(`  Repulsion Force: ${this.repulsionForce}`);
    console.log(`  Spring Constant: ${this.springConstant}`);
    console.log(`  Link Distance: ${this.linkDistance}`);
    console.log(`  Damping: ${this.damping}`);
    console.log("");
  }

  logParticleStates() {
    if (!this.debugMode) {
      return;
    }

    if (this.debugTick >= this.maxDebugTicks) {
      console.log("=== DEBUG MODE COMPLETE ===");
      this.debugMode = false;
    }

    const backend = this.useGPUPhysics ? "GPU" : "CPU";
    console.log(`--- Tick ${this.debugTick} (${backend}) ---`);

    this.nodes.forEach((node, index) => {
      const pos = `(${node.x.toFixed(3)}, ${node.y.toFixed(3)})`;
      const vel = `(${node.vx.toFixed(6)}, ${node.vy.toFixed(6)})`;
      const speed = Math.sqrt(node.vx * node.vx + node.vy * node.vy).toFixed(6);
      console.log(`  ${node.name}: pos=${pos} vel=${vel} speed=${speed}`);
    });

    console.log("");
    this.debugTick++;
  }

  startAnimation() {
    const animate = () => {
      this.updateMessageFlows();
      this.applyForces();
      this.logParticleStates(); // Log after physics update
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
