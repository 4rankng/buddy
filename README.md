# Oncall CLI - Payment Team Dashboard

A sophisticated, extensible command-line interface built with Go and Bubbletea, designed for payment operations teams. This CLI implements a port and adapter architecture pattern with a professional terminal-based GUI.

## Features

### 🎯 Payment Team Dashboard
- **Multi-Pane Interface**: Incident queue, transaction diagnostics, admin actions, and real-time logs
- **Professional TUI**: Built with Bubbletea and styled with Lipgloss for a polished terminal experience
- **Mock Data Integration**: Realistic Jira tickets, transaction flows, and admin actions
- **Real-time Logging**: Timestamped log entries with different severity levels
- **Responsive Navigation**: Tab-based pane switching, arrow key navigation, and keyboard shortcuts

### 🏗️ Architecture
- **Port & Adapter Pattern**: Extensible module system with pluggable components
- **Multi-Pane Design**: Modular pane components for different functionalities
- **Focus Management**: Sophisticated navigation and focus handling system
- **Type Safety**: Strongly typed Go implementation with comprehensive error handling

### 📊 Dashboard Components

#### Incident Queue (Left Pane)
- Mock Jira tickets with status, priority, and assignee information
- Visual status indicators (● ○ ◇ ◐ ✓ ✗)
- Priority color coding (Critical: Red, High: Orange, Medium: Yellow, Low: Green)
- Navigation through incidents with arrow keys

#### Transaction Diagnostics (Top Right)
- Transaction ID input field with cursor management
- Interactive transaction flow visualization
- Mock multi-step transaction processing (Payment → Fraud Check → Authorization → Settlement → Notification)
- Real-time status indicators for each transaction step

#### Admin Actions (Bottom Right)
- Checkbox-based action selection
- Dangerous action warnings
- Module-based action categorization
- Bulk action support

#### Real-time Logs (Bottom Panel)
- Timestamped log entries (HH:MM:SS format)
- Color-coded severity levels (INFO, WARN, ERROR, DEBUG, SUCCESS)
- Scrollable log history
- Auto-scroll to latest entries

## Installation & Usage

### Prerequisites
- Go 1.21 or higher
- Terminal with Unicode support

### Build from Source
```bash
# Clone the repository
git clone <repository-url>
cd oncall-app

# Build the application
make build

# Run the payment team dashboard
./oncall
```

### Available Commands
```bash
# Launch payment team dashboard
./oncall

# Quit application at any time
q or Ctrl+C
```

## Navigation Guide

### Keyboard Shortcuts
- **Tab**: Switch to next pane (Incident → Diagnostics → Admin → Logs)
- **Shift+Tab**: Switch to previous pane
- **↑/↓ Arrow Keys**: Navigate within current pane
- **Enter**: Select/activate current item or action
- **Esc**: Exit input mode or current focus state
- **q**: Quit application

### Pane-Specific Controls

#### Incident Queue
- **↑/↓**: Navigate through incidents
- **Enter**: Open incident details (logs action)
- **Tab**: Switch to next pane

#### Transaction Diagnostics
- **↑/↓**: Navigate through transaction steps (when flow is displayed)
- **Enter**: Focus transaction ID input field
- **f**: Quick access to transaction search
- **In Input Mode**:
  - **Esc**: Exit input mode
  - **Enter**: Execute transaction trace
  - **↑/↓**: Move cursor
  - **Backspace**: Delete character

#### Admin Actions
- **↑/↓**: Navigate through actions
- **Enter**: Toggle action selection (enable/disable)

#### Logs Panel
- **↑/↓**: Scroll through log history
- **Enter**: Scroll to bottom (latest logs)

## Development

### Project Structure
```
oncall-app/
├── cmd/oncall/main.go          # Application entry point
├── pkg/
│   ├── tui/payment/            # Payment team dashboard
│   │   ├── enhanced_dashboard.go    # Main dashboard orchestration
│   │   ├── focus_manager.go         # Navigation and focus management
│   │   ├── layout.go                # Responsive layout engine
│   │   ├── types.go                 # Shared data structures
│   │   └── panes/                   # Modular pane components
│   │       ├── incident_queue.go     # Jira incident management
│   │       ├── diagnostics.go        # Transaction tracing
│   │       ├── admin_actions.go      # Administrative operations
│   │       └── logs.go               # Real-time logging
│   ├── core/                  # Core application logic
│   ├── ports/                 # Interface definitions (extensibility)
│   └── modules/               # Reusable modules
│       ├── doorman/           # SQL query execution
│       ├── jira/              # Ticket management
│       └── datadog/           # Monitoring integration
└── internal/teams/            # Team-specific implementations
    └── payment/               # Payment team workflows
```

### Architecture Patterns

#### Port and Adapter Design
- **Ports**: Interface definitions in `pkg/ports/`
- **Adapters**: Concrete implementations in `pkg/modules/`
- **Clean separation** between business logic and external systems

#### Pane System
- Each pane is a self-contained component with its own state management
- Centralized focus manager handles navigation between panes
- Responsive layout system adapts to terminal size changes

#### Extensibility
- New panes can be added by implementing the pane interface
- New modules (Doorman, Jira, Datadog) follow the adapter pattern
- Team-specific logic can be added in `internal/teams/`

### Contributing

#### Adding New Panes
1. Create a new file in `pkg/tui/payment/panes/`
2. Implement the pane interface with `Render()` method
3. Add the pane to the `EnhancedDashboardModel`
4. Update focus management to include the new pane

#### Adding New Modules
1. Create interface in `pkg/ports/`
2. Implement adapter in `pkg/modules/`
3. Register module in main application
4. Use module in relevant panes

## Technology Stack

- **Language**: Go 1.21+
- **TUI Framework**: Bubbletea (Elm architecture)
- **Styling**: Lipgloss (terminal styling library)
- **Architecture**: Port and Adapter pattern
- **Design**: Component-based, event-driven

## Mock Data

The application includes realistic mock data for demonstration:

### Incidents
- Multiple Jira tickets with different priorities and statuses
- Payment-related issues (double charges, certificate updates, settlement delays)
- Realistic assignees and timestamps

### Transaction Flows
- Multi-step payment processing
- Different transaction statuses (Completed, Processing, Failed, Pending)
- Timing information and system integration points

### Admin Actions
- Payment operations (deregistration, refunds, restart services)
- System management (cache clearing, rate limiting)
- Safety warnings for dangerous operations

## Roadmap

### Phase 1 ✅ (Completed)
- [x] Basic multi-pane dashboard
- [x] Focus management system
- [x] Mock data integration
- [x] Professional styling
- [x] Real-time logging

### Phase 2 🚧 (In Progress)
- [ ] Module integration (real Jira, Datadog, Doorman)
- [ ] Advanced modal dialogs
- [ ] Configuration management
- [ ] Error handling improvements

### Phase 3 📋 (Planned)
- [ ] Multiple team support
- [ ] Plugin system for external modules
- [ ] Performance optimization
- [ ] Comprehensive testing suite

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here]