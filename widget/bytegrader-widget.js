/**
 * ByteGrader Embedded Widget
 * Universal JavaScript widget for any website
 * 
 * Usage:
 * <div id="bytegrader-widget" 
 *      data-api-endpoint="https://your-api.com"
 *      data-assignment-id="hw1"
 *      data-title="Submit Assignment">
 * </div>
 * <script src="path/to/bytegrader-widget.js"></script>
 */

(function() {
    'use strict';
    
    // Prevent multiple initialization
    if (window.ByteGraderWidget) return;
    
    class ByteGraderWidget {
        constructor(element) {
            this.element = element;
            this.config = this.parseConfig(element);
            this.formId = 'bytegrader-' + Math.random().toString(36).substr(2, 9);
            this.pollInterval = null;
            this.timer = null;
            this.startTime = null;
            
            this.init();
        }
        
        parseConfig(element) {
            return {
                apiEndpoint: element.dataset.apiEndpoint || '',
                assignmentId: element.dataset.assignmentId || 'assignment_' + Date.now(),
                title: element.dataset.title || 'Submit Assignment',
                description: element.dataset.description || 'Upload your assignment file for automated grading.',
                maxFileSize: parseInt(element.dataset.maxFileSize) || 50,
                allowedExtensions: (element.dataset.allowedExtensions || 'py,java,cpp,zip,js,c').split(','),
                requireName: element.dataset.requireName === 'true',
                requireEmail: element.dataset.requireEmail === 'true',
                theme: element.dataset.theme || 'default',
                showTimer: element.dataset.showTimer !== 'false',
                maxAttempts: parseInt(element.dataset.maxAttempts) || 0,
                successRedirect: element.dataset.successRedirect || ''
            };
        }
        
        init() {
            if (!this.config.apiEndpoint) {
                this.showConfigError('data-api-endpoint is required');
                return;
            }
            
            this.injectStyles();
            this.renderForm();
            this.attachEventListeners();
            
            // Check API connectivity
            this.checkApiHealth();
        }
        
        showConfigError(message) {
            this.element.innerHTML = `
                <div style="
                    color: #721c24;
                    background: #f8d7da;
                    border: 1px solid #f5c6cb;
                    padding: 20px;
                    border-radius: 4px;
                    font-family: Arial, sans-serif;
                ">
                    <strong>ByteGrader Configuration Error:</strong><br>
                    ${message}
                </div>
            `;
        }
        
        async checkApiHealth() {
            try {
                const response = await fetch(`${this.config.apiEndpoint}/health`, {
                    method: 'GET',
                    timeout: 5000
                });
                
                if (!response.ok) {
                    console.warn('ByteGrader API health check failed:', response.status);
                }
            } catch (error) {
                console.warn('ByteGrader API not accessible:', error.message);
                // Don't show error to user - widget should still be usable
            }
        }
        
        injectStyles() {
            if (document.getElementById('bytegrader-styles')) return;
            
            const styles = `
                .bytegrader-widget {
                    max-width: 600px;
                    margin: 20px 0;
                    padding: 25px;
                    border: 1px solid #e1e5e9;
                    border-radius: 8px;
                    background: #ffffff;
                    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
                    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
                }
                
                .bytegrader-widget h3 {
                    margin: 0 0 10px 0;
                    color: #2c3e50;
                    font-size: 1.4em;
                    font-weight: 600;
                }
                
                .bytegrader-widget .description {
                    margin: 0 0 20px 0;
                    color: #6c757d;
                    line-height: 1.5;
                }
                
                .bytegrader-form-group {
                    margin-bottom: 20px;
                }
                
                .bytegrader-form-group label {
                    display: block;
                    margin-bottom: 6px;
                    font-weight: 500;
                    color: #495057;
                }
                
                .bytegrader-form-group input[type="text"],
                .bytegrader-form-group input[type="email"] {
                    width: 100%;
                    padding: 10px 12px;
                    border: 1px solid #ced4da;
                    border-radius: 4px;
                    font-size: 14px;
                    transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;
                    box-sizing: border-box;
                }
                
                .bytegrader-form-group input[type="text"]:focus,
                .bytegrader-form-group input[type="email"]:focus {
                    outline: none;
                    border-color: #3498db;
                    box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
                }
                
                .bytegrader-file-input {
                    position: relative;
                    display: block;
                    width: 100%;
                    padding: 20px;
                    border: 2px dashed #bdc3c7;
                    border-radius: 6px;
                    background: #f8f9fa;
                    text-align: center;
                    cursor: pointer;
                    transition: all 0.3s ease;
                }
                
                .bytegrader-file-input:hover {
                    border-color: #3498db;
                    background: #e3f2fd;
                }
                
                .bytegrader-file-input.dragover {
                    border-color: #2980b9;
                    background: #bbdefb;
                }
                
                .bytegrader-file-input input[type="file"] {
                    position: absolute;
                    top: 0;
                    left: 0;
                    width: 100%;
                    height: 100%;
                    opacity: 0;
                    cursor: pointer;
                }
                
                .bytegrader-file-text {
                    pointer-events: none;
                }
                
                .bytegrader-file-info {
                    margin-top: 8px;
                    font-size: 0.85em;
                    color: #6c757d;
                }
                
                .bytegrader-file-preview {
                    margin-top: 15px;
                    padding: 12px;
                    background: #e8f5e8;
                    border: 1px solid #c3e6c3;
                    border-radius: 4px;
                    display: none;
                }
                
                .bytegrader-submit-btn {
                    background: #3498db;
                    color: white;
                    border: none;
                    padding: 12px 24px;
                    border-radius: 4px;
                    font-size: 16px;
                    font-weight: 500;
                    cursor: pointer;
                    transition: background-color 0.3s ease;
                    width: 100%;
                }
                
                .bytegrader-submit-btn:hover:not(:disabled) {
                    background: #2980b9;
                }
                
                .bytegrader-submit-btn:disabled {
                    background: #bdc3c7;
                    cursor: not-allowed;
                }
                
                .bytegrader-result {
                    margin-top: 25px;
                    padding: 20px;
                    border-radius: 6px;
                    display: none;
                }
                
                .bytegrader-result.pending {
                    background: #fff3cd;
                    border: 1px solid #ffeaa7;
                    color: #856404;
                }
                
                .bytegrader-result.success {
                    background: #d4edda;
                    border: 1px solid #c3e6cb;
                    color: #155724;
                }
                
                .bytegrader-result.error {
                    background: #f8d7da;
                    border: 1px solid #f5c6cb;
                    color: #721c24;
                }
                
                .bytegrader-spinner {
                    display: inline-block;
                    width: 20px;
                    height: 20px;
                    border: 2px solid #f3f3f3;
                    border-top: 2px solid #3498db;
                    border-radius: 50%;
                    animation: bytegrader-spin 1s linear infinite;
                    margin-right: 10px;
                }
                
                @keyframes bytegrader-spin {
                    0% { transform: rotate(0deg); }
                    100% { transform: rotate(360deg); }
                }
                
                .bytegrader-score {
                    font-size: 1.2em;
                    font-weight: 600;
                    margin: 10px 0;
                }
                
                .bytegrader-feedback {
                    margin: 15px 0;
                    padding: 15px;
                    background: #f8f9fa;
                    border-radius: 4px;
                    white-space: pre-wrap;
                    font-family: monospace;
                    font-size: 0.9em;
                    line-height: 1.4;
                }
                
                .bytegrader-timer {
                    margin-top: 10px;
                    font-size: 0.9em;
                    color: #6c757d;
                }
                
                .bytegrader-retry-btn {
                    background: #dc3545;
                    color: white;
                    border: none;
                    padding: 8px 16px;
                    border-radius: 4px;
                    cursor: pointer;
                    margin-top: 10px;
                }
                
                .bytegrader-retry-btn:hover {
                    background: #c82333;
                }
            `;
            
            const styleSheet = document.createElement('style');
            styleSheet.id = 'bytegrader-styles';
            styleSheet.textContent = styles;
            document.head.appendChild(styleSheet);
        }
        
        renderForm() {
            const studentFieldsHtml = `
                ${this.config.requireName ? `
                    <div class="bytegrader-form-group">
                        <label for="${this.formId}-name">Name *</label>
                        <input type="text" id="${this.formId}-name" name="student_name" required>
                    </div>
                ` : ''}
                ${this.config.requireEmail ? `
                    <div class="bytegrader-form-group">
                        <label for="${this.formId}-email">Email *</label>
                        <input type="email" id="${this.formId}-email" name="student_email" required>
                    </div>
                ` : ''}
            `;
            
            this.element.innerHTML = `
                <div class="bytegrader-widget">
                    <h3>${this.config.title}</h3>
                    <p class="description">${this.config.description}</p>
                    
                    <form id="${this.formId}" class="bytegrader-form">
                        ${studentFieldsHtml}
                        
                        <div class="bytegrader-form-group">
                            <label>Assignment File *</label>
                            <div class="bytegrader-file-input">
                                <input type="file" id="${this.formId}-file" name="file" required 
                                       accept="${this.getAcceptString()}">
                                <div class="bytegrader-file-text">
                                    <div>📁 Click to choose file or drag and drop</div>
                                    <div class="bytegrader-file-info">
                                        Allowed: ${this.config.allowedExtensions.join(', ').toUpperCase()} 
                                        (Max: ${this.config.maxFileSize}MB)
                                    </div>
                                </div>
                            </div>
                            <div class="bytegrader-file-preview" id="${this.formId}-preview"></div>
                        </div>
                        
                        <div class="bytegrader-form-group">
                            <button type="submit" class="bytegrader-submit-btn">
                                Submit for Grading
                            </button>
                        </div>
                    </form>
                    
                    <div id="${this.formId}-result" class="bytegrader-result"></div>
                </div>
            `;
        }
        
        getAcceptString() {
            return this.config.allowedExtensions.map(ext => '.' + ext).join(',');
        }
        
        attachEventListeners() {
            const form = document.getElementById(this.formId);
            const fileInput = document.getElementById(`${this.formId}-file`);
            const fileInputContainer = fileInput.parentElement;
            const preview = document.getElementById(`${this.formId}-preview`);
            
            // Form submission
            form.addEventListener('submit', (e) => this.handleSubmit(e));
            
            // File input change
            fileInput.addEventListener('change', (e) => this.handleFileSelect(e));
            
            // Drag and drop
            fileInputContainer.addEventListener('dragover', (e) => {
                e.preventDefault();
                fileInputContainer.classList.add('dragover');
            });
            
            fileInputContainer.addEventListener('dragleave', (e) => {
                e.preventDefault();
                fileInputContainer.classList.remove('dragover');
            });
            
            fileInputContainer.addEventListener('drop', (e) => {
                e.preventDefault();
                fileInputContainer.classList.remove('dragover');
                
                const files = e.dataTransfer.files;
                if (files.length > 0) {
                    fileInput.files = files;
                    this.handleFileSelect({ target: fileInput });
                }
            });
        }
        
        handleFileSelect(e) {
            const file = e.target.files[0];
            const preview = document.getElementById(`${this.formId}-preview`);
            
            if (!file) {
                preview.style.display = 'none';
                return;
            }
            
            // Validate file
            if (!this.validateFile(file)) {
                e.target.value = '';
                preview.style.display = 'none';
                return;
            }
            
            // Show preview
            preview.innerHTML = `
                <div>✅ <strong>${file.name}</strong></div>
                <div>Size: ${this.formatFileSize(file.size)}</div>
            `;
            preview.style.display = 'block';
        }
        
        validateFile(file) {
            // Check size
            const maxBytes = this.config.maxFileSize * 1024 * 1024;
            if (file.size > maxBytes) {
                alert(`File too large. Maximum size is ${this.config.maxFileSize}MB.`);
                return false;
            }
            
            // Check extension
            const fileName = file.name.toLowerCase();
            const extension = fileName.split('.').pop();
            
            if (!this.config.allowedExtensions.includes(extension)) {
                alert(`File type not allowed. Please use: ${this.config.allowedExtensions.join(', ')}`);
                return false;
            }
            
            return true;
        }
        
        formatFileSize(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }
        
        async handleSubmit(e) {
            e.preventDefault();
            
            const form = e.target;
            const submitBtn = form.querySelector('.bytegrader-submit-btn');
            
            // Disable form
            submitBtn.disabled = true;
            submitBtn.textContent = 'Submitting...';
            
            // Show pending status
            this.showPendingStatus();
            
            try {
                // Create form data
                const formData = new FormData(form);
                
                // Submit to API
                const response = await fetch(`${this.config.apiEndpoint}/submit`, {
                    method: 'POST',
                    body: formData
                });
                
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                
                const data = await response.json();
                
                if (data.job_id) {
                    // Start polling for results
                    this.startPolling(data.job_id);
                } else {
                    throw new Error('No job ID returned from API');
                }
                
            } catch (error) {
                console.error('Submission error:', error);
                this.showError(`Submission failed: ${error.message}`);
                this.resetForm();
            }
        }
        
        showPendingStatus() {
            const resultDiv = document.getElementById(`${this.formId}-result`);
            
            resultDiv.className = 'bytegrader-result pending';
            resultDiv.innerHTML = `
                <div style="display: flex; align-items: center;">
                    <div class="bytegrader-spinner"></div>
                    <div>
                        <div><strong>Grading in Progress</strong></div>
                        <div>Your submission is being processed. This may take a few minutes...</div>
                        ${this.config.showTimer ? `<div class="bytegrader-timer">Time elapsed: <span id="${this.formId}-timer">0:00</span></div>` : ''}
                    </div>
                </div>
            `;
            resultDiv.style.display = 'block';
            
            // Start timer
            if (this.config.showTimer) {
                this.startTimer();
            }
        }
        
        startTimer() {
            this.startTime = Date.now();
            this.timer = setInterval(() => {
                const elapsed = Math.floor((Date.now() - this.startTime) / 1000);
                const minutes = Math.floor(elapsed / 60);
                const seconds = elapsed % 60;
                const timerElement = document.getElementById(`${this.formId}-timer`);
                if (timerElement) {
                    timerElement.textContent = `${minutes}:${seconds.toString().padStart(2, '0')}`;
                }
            }, 1000);
        }
        
        stopTimer() {
            if (this.timer) {
                clearInterval(this.timer);
                this.timer = null;
            }
        }
        
        async startPolling(jobId) {
            const maxAttempts = 60; // 5 minutes
            let attempts = 0;
            
            this.pollInterval = setInterval(async () => {
                attempts++;
                
                if (attempts > maxAttempts) {
                    clearInterval(this.pollInterval);
                    this.showError('Grading timeout. Please contact your instructor.');
                    this.resetForm();
                    return;
                }
                
                try {
                    const response = await fetch(`${this.config.apiEndpoint}/status/${jobId}`);
                    
                    if (!response.ok) {
                        throw new Error(`HTTP ${response.status}`);
                    }
                    
                    const data = await response.json();
                    
                    if (data.job && data.job.status === 'completed') {
                        clearInterval(this.pollInterval);
                        this.showSuccess(data.job.result);
                    } else if (data.job && data.job.status === 'failed') {
                        clearInterval(this.pollInterval);
                        this.showError(data.job.result ? data.job.result.error : 'Grading failed');
                        this.resetForm();
                    }
                    // Continue polling if still processing
                    
                } catch (error) {
                    console.error('Polling error:', error);
                    clearInterval(this.pollInterval);
                    this.showError('Failed to check grading status');
                    this.resetForm();
                }
            }, 5000); // Poll every 5 seconds
        }
        
        showSuccess(result) {
            this.stopTimer();
            
            const resultDiv = document.getElementById(`${this.formId}-result`);
            
            resultDiv.className = 'bytegrader-result success';
            resultDiv.innerHTML = `
                <div><strong>✅ Grading Complete!</strong></div>
                <div class="bytegrader-score">Score: ${result.score}/100</div>
                <div><strong>Feedback:</strong></div>
                <div class="bytegrader-feedback">${result.feedback}</div>
            `;
            
            // Handle success redirect
            if (this.config.successRedirect) {
                setTimeout(() => {
                    window.location.href = this.config.successRedirect;
                }, 3000);
            }
            
            // Reset form for another submission
            setTimeout(() => this.resetForm(), 5000);
        }
        
        showError(message) {
            this.stopTimer();
            
            const resultDiv = document.getElementById(`${this.formId}-result`);
            
            resultDiv.className = 'bytegrader-result error';
            resultDiv.innerHTML = `
                <div><strong>❌ Error</strong></div>
                <div>${message}</div>
                <div style="margin-top: 10px;">
                    <button class="bytegrader-retry-btn" onclick="location.reload()">
                        Try Again
                    </button>
                </div>
            `;
        }
        
        resetForm() {
            const form = document.getElementById(this.formId);
            const submitBtn = form.querySelector('.bytegrader-submit-btn');
            
            submitBtn.disabled = false;
            submitBtn.textContent = 'Submit for Grading';
            
            // Clear file preview
            const preview = document.getElementById(`${this.formId}-preview`);
            preview.style.display = 'none';
        }
    }
    
    // Initialize all ByteGrader widgets on the page
    function initByteGraderWidgets() {
        const widgets = document.querySelectorAll('[id^="bytegrader-widget"], [data-api-endpoint]');
        widgets.forEach(widget => {
            if (!widget.dataset.initialized) {
                new ByteGraderWidget(widget);
                widget.dataset.initialized = 'true';
            }
        });
    }
    
    // Auto-initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initByteGraderWidgets);
    } else {
        initByteGraderWidgets();
    }
    
    // Expose for manual initialization
    window.ByteGraderWidget = ByteGraderWidget;
    window.initByteGraderWidgets = initByteGraderWidgets;
    
})();
