class Dashboard {
    constructor(config) {
        this.config = {
            apiEndpoint: config.apiEndpoint || '/api/v1',
            websocketEndpoint: config.websocketEndpoint || '/ws',
            refreshInterval: (config.refreshInterval || 30) * 1000,
            auth: config.auth || false
        };
        
        this.websocket = null;
        this.charts = {};
        this.connected = false;
        this.currentSection = 'overview';
        this.refreshTimer = null;
    }

    init() {
        this.setupEventListeners();
        this.setupWebSocket();
        this.loadOverview();
        this.startRefreshTimer();
        
        console.log('Dashboard initialized');
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-link[data-section]').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = e.target.closest('.nav-link').dataset.section;
                this.showSection(section);
            });
        });

        // Window resize
        window.addEventListener('resize', () => {
            this.resizeCharts();
        });
    }

    setupWebSocket() {
        const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${location.host}${this.websocketEndpoint}`;
        
        try {
            this.websocket = new WebSocket(wsUrl);
            
            this.websocket.onopen = () => {
                this.connected = true;
                this.updateConnectionStatus('connected');
                console.log('WebSocket connected');
            };
            
            this.websocket.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleWebSocketMessage(message);
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };
            
            this.websocket.onclose = () => {
                this.connected = false;
                this.updateConnectionStatus('disconnected');
                console.log('WebSocket disconnected');
                
                // Attempt to reconnect after 5 seconds
                setTimeout(() => {
                    this.setupWebSocket();
                }, 5000);
            };
            
            this.websocket.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.updateConnectionStatus('error');
            };
        } catch (error) {
            console.error('Error setting up WebSocket:', error);
            this.updateConnectionStatus('error');
        }
    }

    handleWebSocketMessage(message) {
        switch (message.type) {
            case 'realtime_update':
                this.updateRealtimeData(message.data);
                break;
            case 'alert':
                this.handleAlert(message.data);
                break;
            default:
                console.log('Unknown message type:', message.type);
        }
    }

    updateRealtimeData(data) {
        this.updateOverviewMetrics(data);
        this.updateSystemMetrics(data.system_metrics);
        this.updatePerformanceMetrics(data.performance);
        this.updateErrorMetrics(data.errors);
        this.updateTrafficMetrics(data.traffic);
        this.updateLastUpdateTime();
    }

    updateConnectionStatus(status) {
        const statusElement = document.getElementById('connection-status');
        const statusIcon = statusElement.parentElement.querySelector('i');
        
        statusIcon.className = 'fas fa-circle me-1';
        
        switch (status) {
            case 'connected':
                statusIcon.classList.add('text-success');
                statusElement.textContent = 'Connected';
                break;
            case 'disconnected':
                statusIcon.classList.add('text-danger');
                statusElement.textContent = 'Disconnected';
                break;
            case 'connecting':
                statusIcon.classList.add('text-warning');
                statusElement.textContent = 'Connecting...';
                break;
            case 'error':
                statusIcon.classList.add('text-danger');
                statusElement.textContent = 'Connection Error';
                break;
        }
    }

    updateLastUpdateTime() {
        const element = document.getElementById('last-update');
        if (element) {
            element.textContent = new Date().toLocaleTimeString();
        }
    }

    showSection(sectionName) {
        // Update navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        document.querySelector(`[data-section="${sectionName}"]`).classList.add('active');
        
        // Hide all sections
        document.querySelectorAll('.dashboard-section').forEach(section => {
            section.style.display = 'none';
        });
        
        // Show selected section
        const targetSection = document.getElementById(`${sectionName}-section`);
        if (targetSection) {
            targetSection.style.display = 'block';
            targetSection.classList.add('fade-in');
        }
        
        this.currentSection = sectionName;
        
        // Load section data
        this.loadSectionData(sectionName);
    }

    loadSectionData(section) {
        switch (section) {
            case 'overview':
                this.loadOverview();
                break;
            case 'system':
                this.loadSystemMetrics();
                break;
            case 'performance':
                this.loadPerformanceMetrics();
                break;
            case 'business':
                this.loadBusinessMetrics();
                break;
            case 'features':
                this.loadFeatureFlags();
                break;
            case 'cache':
                this.loadCacheStats();
                break;
            case 'ratelimit':
                this.loadRateLimitStats();
                break;
            case 'tracing':
                this.loadTracingData();
                break;
            case 'alerts':
                this.loadAlerts();
                break;
        }
    }

    async loadOverview() {
        try {
            const response = await this.apiCall('/overview');
            const data = response;
            
            this.updateOverviewCards(data.metrics);
            this.updateSystemHealth(data.system_health);
            this.updateComponentStatus(data.components);
            this.updateAlertsBadge(data.alerts);
            
            this.createResponseTimeChart();
            this.createTrafficSourcesChart();
        } catch (error) {
            console.error('Error loading overview:', error);
        }
    }

    updateOverviewCards(metrics) {
        if (!metrics) return;
        
        this.updateElement('system-health', `${metrics.overall_score || 100}%`);
        this.updateElement('request-rate', this.formatNumber(metrics.request_rate || 0));
        this.updateElement('error-rate', `${(metrics.error_rate || 0).toFixed(2)}%`);
        this.updateElement('active-sessions', this.formatNumber(metrics.active_sessions || 0));
    }

    updateSystemHealth(health) {
        if (!health) return;
        
        const healthCard = document.querySelector('.bg-success');
        const statusElement = document.getElementById('system-status');
        
        // Update card color based on status
        healthCard.className = healthCard.className.replace(/bg-\w+/, this.getHealthColor(health.status));
        
        if (statusElement) {
            statusElement.textContent = health.status === 'healthy' ? 'All systems operational' : 
                                        health.status === 'degraded' ? 'Some issues detected' : 
                                        'Critical issues detected';
        }
    }

    getHealthColor(status) {
        switch (status) {
            case 'healthy': return 'bg-success';
            case 'degraded': return 'bg-warning';
            case 'unhealthy': return 'bg-danger';
            default: return 'bg-secondary';
        }
    }

    updateComponentStatus(components) {
        if (!components) return;
        
        const container = document.getElementById('components-status');
        container.innerHTML = '';
        
        components.forEach(component => {
            const componentHtml = `
                <div class="col-xl-4 col-md-6 mb-3">
                    <div class="component-status ${component.status}">
                        <div class="d-flex justify-content-between align-items-start">
                            <div>
                                <div class="component-name">${component.name}</div>
                                <div class="component-message">${component.message || ''}</div>
                                <small class="text-muted">Last check: ${this.formatTime(component.last_check)}</small>
                            </div>
                            <div class="text-end">
                                <div class="component-health text-${this.getStatusColor(component.status)}">
                                    ${component.health.toFixed(1)}%
                                </div>
                                <small class="text-muted text-uppercase">${component.status}</small>
                            </div>
                        </div>
                    </div>
                </div>
            `;
            container.innerHTML += componentHtml;
        });
    }

    getStatusColor(status) {
        switch (status) {
            case 'healthy': return 'success';
            case 'degraded': return 'warning';
            case 'unhealthy': return 'danger';
            default: return 'secondary';
        }
    }

    updateAlertsBadge(alerts) {
        const badge = document.getElementById('alert-count');
        if (alerts && alerts.length > 0) {
            badge.textContent = alerts.length;
            badge.style.display = 'inline';
        } else {
            badge.style.display = 'none';
        }
    }

    async loadAlerts() {
        try {
            const alerts = await this.apiCall('/alerts');
            this.displayAlerts(alerts);
        } catch (error) {
            console.error('Error loading alerts:', error);
        }
    }

    displayAlerts(alerts) {
        const container = document.getElementById('alerts-list');
        
        if (!alerts || alerts.length === 0) {
            container.innerHTML = '<div class="text-center text-muted">No active alerts</div>';
            return;
        }
        
        container.innerHTML = '';
        alerts.forEach(alert => {
            const alertHtml = `
                <div class="alert-item alert-${alert.severity}">
                    <div class="d-flex justify-content-between align-items-start">
                        <div class="flex-grow-1">
                            <div class="d-flex align-items-center mb-2">
                                <span class="alert-severity ${alert.severity}">${alert.severity}</span>
                                <span class="alert-title ms-2">${alert.title}</span>
                            </div>
                            <div class="alert-description">${alert.description}</div>
                            <div class="alert-meta">
                                <small>
                                    <i class="fas fa-clock me-1"></i>
                                    ${this.formatTime(alert.created_at)}
                                    ${alert.component ? `• <i class="fas fa-server me-1"></i>${alert.component}` : ''}
                                </small>
                            </div>
                            ${alert.actions ? this.renderAlertActions(alert.actions) : ''}
                        </div>
                    </div>
                </div>
            `;
            container.innerHTML += alertHtml;
        });
    }

    renderAlertActions(actions) {
        let actionsHtml = '<div class="alert-actions">';
        actions.forEach(action => {
            actionsHtml += `<button class="btn btn-outline-primary btn-sm me-2">${action.label}</button>`;
        });
        actionsHtml += '</div>';
        return actionsHtml;
    }

    createResponseTimeChart() {
        const ctx = document.getElementById('response-time-chart');
        if (!ctx) return;
        
        if (this.charts.responseTime) {
            this.charts.responseTime.destroy();
        }
        
        this.charts.responseTime = new Chart(ctx, {
            type: 'line',
            data: {
                labels: this.generateTimeLabels(),
                datasets: [{
                    label: 'Average Response Time',
                    data: this.generateMockData(20, 30, 80),
                    borderColor: '#0d6efd',
                    backgroundColor: 'rgba(13, 110, 253, 0.1)',
                    fill: true,
                    tension: 0.4
                }, {
                    label: 'P95 Response Time',
                    data: this.generateMockData(20, 50, 120),
                    borderColor: '#fd7e14',
                    backgroundColor: 'rgba(253, 126, 20, 0.1)',
                    fill: false,
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'top'
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Response Time (ms)'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    }
                },
                interaction: {
                    intersect: false,
                    mode: 'index'
                }
            }
        });
    }

    createTrafficSourcesChart() {
        const ctx = document.getElementById('traffic-sources-chart');
        if (!ctx) return;
        
        if (this.charts.trafficSources) {
            this.charts.trafficSources.destroy();
        }
        
        this.charts.trafficSources = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Web', 'Mobile', 'API'],
                datasets: [{
                    data: [65, 25, 10],
                    backgroundColor: ['#0d6efd', '#198754', '#ffc107']
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: true,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                }
            }
        });
    }

    generateTimeLabels(count = 20) {
        const labels = [];
        const now = new Date();
        for (let i = count - 1; i >= 0; i--) {
            const time = new Date(now - i * 60000); // 1 minute intervals
            labels.push(time.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}));
        }
        return labels;
    }

    generateMockData(count, min, max) {
        const data = [];
        for (let i = 0; i < count; i++) {
            data.push(Math.floor(Math.random() * (max - min + 1)) + min);
        }
        return data;
    }

    async apiCall(endpoint) {
        const url = `${this.config.apiEndpoint}${endpoint}`;
        const headers = {};
        
        if (this.config.auth) {
            headers['Authorization'] = `Bearer ${localStorage.getItem('dashboard_token')}`;
        }
        
        const response = await fetch(url, { headers });
        
        if (!response.ok) {
            throw new Error(`API call failed: ${response.statusText}`);
        }
        
        return await response.json();
    }

    updateElement(id, value) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = value;
        }
    }

    formatNumber(num) {
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        }
        if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K';
        }
        return num.toString();
    }

    formatTime(timeString) {
        try {
            const date = new Date(timeString);
            return date.toLocaleString();
        } catch (error) {
            return timeString;
        }
    }

    resizeCharts() {
        Object.values(this.charts).forEach(chart => {
            if (chart && typeof chart.resize === 'function') {
                chart.resize();
            }
        });
    }

    startRefreshTimer() {
        if (this.refreshTimer) {
            clearInterval(this.refreshTimer);
        }
        
        this.refreshTimer = setInterval(() => {
            if (this.connected && this.currentSection === 'overview') {
                this.loadOverview();
            }
        }, this.config.refreshInterval);
    }

    // Placeholder methods for other sections
    async loadSystemMetrics() {
        console.log('Loading system metrics...');
    }

    async loadPerformanceMetrics() {
        console.log('Loading performance metrics...');
    }

    async loadBusinessMetrics() {
        console.log('Loading business metrics...');
    }

    async loadFeatureFlags() {
        console.log('Loading feature flags...');
    }

    async loadCacheStats() {
        console.log('Loading cache stats...');
    }

    async loadRateLimitStats() {
        console.log('Loading rate limit stats...');
    }

    async loadTracingData() {
        console.log('Loading tracing data...');
    }

    updateOverviewMetrics(data) {
        // Update overview metrics from real-time data
        if (data && data.performance) {
            this.updateElement('request-rate', this.formatNumber(data.performance.request_rate));
        }
        if (data && data.errors) {
            this.updateElement('error-rate', `${(data.errors.error_rate || 0).toFixed(2)}%`);
        }
        if (data && data.traffic) {
            this.updateElement('active-sessions', this.formatNumber(data.traffic.active_sessions));
        }
    }

    updateSystemMetrics(systemMetrics) {
        // Update system metrics displays
        console.log('Updating system metrics:', systemMetrics);
    }

    updatePerformanceMetrics(performance) {
        // Update performance metrics displays
        console.log('Updating performance metrics:', performance);
    }

    updateErrorMetrics(errors) {
        // Update error metrics displays
        console.log('Updating error metrics:', errors);
    }

    updateTrafficMetrics(traffic) {
        // Update traffic metrics displays
        console.log('Updating traffic metrics:', traffic);
    }

    handleAlert(alertData) {
        // Handle incoming alerts
        console.log('New alert:', alertData);
        
        // Show notification or update alerts section
        if (this.currentSection === 'alerts') {
            this.loadAlerts();
        }
    }
}