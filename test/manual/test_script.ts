interface HeartbeatMessage {
  version: number;
  timestamp: string;
  pid: number;
}

class TestApp {
  private version: number = 10;
  private readonly startTime: Date;

  constructor() { 
    this.startTime = new Date();
  }

  private getHeartbeat(): HeartbeatMessage {
    return {
      version: this.version,
      timestamp: new Date().toISOString(),
      pid: process.pid,
    };
  }
 
  public start(): void {
    console.log("TypeScript Test script started");
    console.log("Start time:", this.startTime.toISOString());
    console.log("Process ID:", process.pid);
    console.log("Version:", this.version);

    // Keep the process running with typed heartbeat
    setInterval(() => {
      const heartbeat = this.getHeartbeat();
      console.log(
        `Heartbeat - Version: ${heartbeat.version} [PID: ${heartbeat.pid}]`
      );
    }, 2000);
  }
}

// Start the application
new TestApp().start();
