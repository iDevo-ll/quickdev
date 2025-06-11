console.log("Test script started");
console.log("Current time:", new Date().toISOString());
console.log("Process ID:", process.pid);

// This variable will be changed to test hot reloading
let version = 3;
console.log("Version:", version);

// Keep the process running
setInterval(() => {
  console.log("Heartbeat - Version:", version);
}, 2000); 
