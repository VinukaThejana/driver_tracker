# Driver Tracker

## Overview

Driver Tracker is an innovative solution for tracking driver locations in real-time. It allows passengers to view and share live locations, while drivers, admins, and passengers can access detailed location histories from pickup to drop-off. The system also includes automated job management to ensure efficiency and compliance.

## Features

- **Real-Time Location Tracking**: Passengers can view and share the live location of their driver.
- **Location History Logging**: Comprehensive logging of the driver's route from acceptance to drop-off.
- **Automated Job Termination**: Automatically end driver jobs that exceed 24 hours.
- **Periodic Location Updates**: Regular interval updates for precise tracking.
- **Real-Time Communication**: Seamless updates with WebSockets for real-time location streaming.

## How It Works

1. **Live Location Sharing**:
   - Passengers can see the driver's real-time location on a map and share it with others.
   
2. **Location History**:
   - The system logs the driver's route, accessible by drivers, admins, and passengers for transparency and record-keeping.

3. **Job Management**:
   - Cron jobs ensure that driver jobs are automatically terminated if they exceed a 24-hour limit, maintaining operational efficiency.

4. **Real-Time Updates**:
   - WebSockets enable real-time streaming of the driver's location, providing passengers with up-to-the-minute updates.

## Usage Scenarios

- **Passenger Tracking**: Passengers can track their ride in real-time and share their driver's location with friends and family for added safety.
- **Driver Accountability**: Drivers have access to their location history, ensuring transparency and accountability.
- **Admin Oversight**: Admins can monitor driver activities and manage job durations effectively.

## Contribution

Contributions are welcome! If you have any ideas, suggestions, or improvements, feel free to open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

---

Enhance your ride-hailing experience with Driver Tracker. Accurate, reliable, and real-time!
