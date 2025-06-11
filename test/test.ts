console.log("Hello from test script!");

// This is a test script that will be reloaded when changed
let counter = 0;

setInterval(() => {
  counter++;
  console.log(`Counter: ${counter}`);
}, 1000);
