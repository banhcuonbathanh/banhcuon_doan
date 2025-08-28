# Enhanced ShopEasy Logger System - Complete Implementation

## 📋 Overview of Enhancements

This enhanced logger system includes:

1. **Enhanced Configuration Layer** with environment variables, hot reload, and schema validation
2. **Async Processing** with log buffering and batch operations
3. **Observability Integration** with metrics, health checks, and distributed tracing
4. **Multiple Output Destinations** (console, files, external services)
5. **Specialized Loggers** for audit, security, and performance
6. **Visual Improvements** with color coding and descriptive labels

## 🔄 Step 2: Complete Async Processing Implementation

### 2.1 Complete Log Buffer Implementation
`internal/logger/async/buffer.go`:

```go
package async

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
)

var (
    ErrBufferFull    = fmt.Errorf("log buffer is full")
    ErrBufferTimeout = fmt.Errorf("log buffer timeout")
    ErrBufferClosed  = fmt.Errorf("log buffer is closed")
)

// LogBuffer manages buffering of log entries for async processing
type LogBuffer struct {
    entries  chan *core.LogEntry
    done     chan struct{}
    wg       sync.WaitGroup
    maxSize  int
    timeout  time.Duration
    metrics  *BufferMetrics
    mu       sync.RWMutex
    closed   int32
}

type BufferMetrics struct {
    EntriesReceived    int64
    EntriesProcessed   int64
    EntriesDropped     int64
    BufferUtilization  float64
    AverageWaitTime    time.Duration
    LastFlushTime      time.Time
    FlushCount         int64
    mu                 sync.RWMutex
}

func NewLogBuffer(maxSize int, timeout time.Duration) *LogBuffer {
    return &LogBuffer{
        entries: make(chan *core.LogEntry, maxSize),
        done:    make(chan struct{}),
        maxSize: maxSize,
        timeout: timeout,
        metrics: &BufferMetrics{},
    }
}

func (lb *LogBuffer) Add(entry *core.LogEntry) error {
    if atomic.LoadInt32(&lb.closed) == 1 {
        return ErrBufferClosed
    }
    
    start := time.Now()
    
    select {
    case lb.entries <- entry:
        lb.updateMetrics(func(m *BufferMetrics) {
            atomic.AddInt64(&m.EntriesReceived, 1)
            m.BufferUtilization = float64(len(lb.entries)) / float64(lb.maxSize)
            m.AverageWaitTime = time.Since(start)
        })
        return nil
    case <-time.After(lb.timeout):
        lb.updateMetrics(func(m *BufferMetrics) {
            atomic.AddInt64(&m.EntriesDropped, 1)
        })
        return ErrBufferTimeout
    case <-lb.done:
        return ErrBufferClosed
    default:
        lb.updateMetrics(func(m *BufferMetrics) {
            atomic.AddInt64(&m.EntriesDropped, 1)
        })
        return ErrBufferFull
    }
}

func (lb *LogBuffer) Consume(ctx context.Context, batchSize int, processor func([]*core.LogEntry) error) {
    lb.wg.Add(1)
    defer lb.wg.Done()
    
    batch := make([]*core.LogEntry, 0, batchSize)
    ticker := time.NewTicker(time.Second) // Flush every second if batch not full
    defer ticker.Stop()
    
    for {
        select {
        case entry, ok := <-lb.entries:
            if !ok {
                // Channel closed, process remaining batch
                if len(batch) > 0 {
                    lb.processBatch(batch, processor)
                }
                return
            }
            
            batch = append(batch, entry)
            
            if len(batch) >= batchSize {
                lb.processBatch(batch, processor)
                batch = batch[:0] // Reset slice
            }
            
        case <-ticker.C:
            if len(batch) > 0 {
                lb.processBatch(batch, processor)
                batch = batch[:0]
            }
            
        case <-ctx.Done():
            return
            
        case <-lb.done:
            if len(batch) > 0 {
                lb.processBatch(batch, processor)
            }
            return
        }
    }
}

func (lb *LogBuffer) processBatch(batch []*core.LogEntry, processor func([]*core.LogEntry) error) {
    start := time.Now()
    
    if err := processor(batch); err != nil {
        // Log processing error (but avoid infinite loop)
        fmt.Printf("Error processing log batch: %v\n", err)
    }
    
    lb.updateMetrics(func(m *BufferMetrics) {
        atomic.AddInt64(&m.EntriesProcessed, int64(len(batch)))
        atomic.AddInt64(&m.FlushCount, 1)
        m.LastFlushTime = start
    })
}

func (lb *LogBuffer) Close() error {
    if atomic.CompareAndSwapInt32(&lb.closed, 0, 1) {
        close(lb.done)
        close(lb.entries)
        lb.wg.Wait()
    }
    return nil
}

func (lb *LogBuffer) updateMetrics(fn func(*BufferMetrics)) {
    lb.mu.Lock()
    defer lb.mu.Unlock()
    fn(lb.metrics)
}

func (lb *LogBuffer) GetMetrics() BufferMetrics {
    lb.mu.RLock()
    defer lb.mu.RUnlock()
    
    return BufferMetrics{
        EntriesReceived:   atomic.LoadInt64(&lb.metrics.EntriesReceived),
        EntriesProcessed:  atomic.LoadInt64(&lb.metrics.EntriesProcessed),
        EntriesDropped:    atomic.LoadInt64(&lb.metrics.EntriesDropped),
        BufferUtilization: lb.metrics.BufferUtilization,
        AverageWaitTime:   lb.metrics.AverageWaitTime,
        LastFlushTime:     lb.metrics.LastFlushTime,
        FlushCount:        atomic.LoadInt64(&lb.metrics.FlushCount),
    }
}
```

### 2.2 Batch Processing Implementation
`internal/logger/async/batch.go`:

```go
package async

import (
    "context"
    "sync"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
)

type BatchProcessor struct {
    batchSize     int
    flushInterval time.Duration
    processor     func([]*core.LogEntry) error
    buffer        []*core.LogEntry
    mu            sync.Mutex
    lastFlush     time.Time
}

func NewBatchProcessor(batchSize int, flushInterval time.Duration, processor func([]*core.LogEntry) error) *BatchProcessor {
    return &BatchProcessor{
        batchSize:     batchSize,
        flushInterval: flushInterval,
        processor:     processor,
        buffer:        make([]*core.LogEntry, 0, batchSize),
        lastFlush:     time.Now(),
    }
}

func (bp *BatchProcessor) Add(entry *core.LogEntry) error {
    bp.mu.Lock()
    defer bp.mu.Unlock()
    
    bp.buffer = append(bp.buffer, entry)
    
    // Check if we should flush
    shouldFlush := len(bp.buffer) >= bp.batchSize ||
        time.Since(bp.lastFlush) >= bp.flushInterval
    
    if shouldFlush {
        return bp.flushLocked()
    }
    
    return nil
}

func (bp *BatchProcessor) Flush() error {
    bp.mu.Lock()
    defer bp.mu.Unlock()
    return bp.flushLocked()
}

func (bp *BatchProcessor) flushLocked() error {
    if len(bp.buffer) == 0 {
        return nil
    }
    
    batch := make([]*core.LogEntry, len(bp.buffer))
    copy(batch, bp.buffer)
    bp.buffer = bp.buffer[:0]
    bp.lastFlush = time.Now()
    
    return bp.processor(batch)
}

func (bp *BatchProcessor) StartPeriodicFlush(ctx context.Context) {
    ticker := time.NewTicker(bp.flushInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            bp.Flush()
        case <-ctx.Done():
            bp.Flush() // Final flush
            return
        }
    }
}
```

### 2.3 Worker Pool Implementation
`internal/logger/async/worker.go`:

```go
package async

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
)

type WorkerPool struct {
    workerCount   int
    workQueue     chan *core.LogEntry
    processor     func(*core.LogEntry) error
    workers       []*Worker
    wg            sync.WaitGroup
    ctx           context.Context
    cancel        context.CancelFunc
    metrics       *WorkerMetrics
}

type Worker struct {
    id        int
    workQueue chan *core.LogEntry
    processor func(*core.LogEntry) error
    metrics   *WorkerMetrics
}

type WorkerMetrics struct {
    TotalJobs       int64
    CompletedJobs   int64
    FailedJobs      int64
    AverageTime     time.Duration
    ActiveWorkers   int32
    QueueSize       int32
}

func NewWorkerPool(workerCount int, queueSize int, processor func(*core.LogEntry) error) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &WorkerPool{
        workerCount: workerCount,
        workQueue:   make(chan *core.LogEntry, queueSize),
        processor:   processor,
        ctx:         ctx,
        cancel:      cancel,
        metrics:     &WorkerMetrics{},
    }
}

func (wp *WorkerPool) Start() {
    wp.workers = make([]*Worker, wp.workerCount)
    
    for i := 0; i < wp.workerCount; i++ {
        worker := &Worker{
            id:        i,
            workQueue: wp.workQueue,
            processor: wp.processor,
            metrics:   wp.metrics,
        }
        wp.workers[i] = worker
        
        wp.wg.Add(1)
        go wp.runWorker(worker)
    }
}

func (wp *WorkerPool) Submit(entry *core.LogEntry) error {
    select {
    case wp.workQueue <- entry:
        atomic.AddInt64(&wp.metrics.TotalJobs, 1)
        atomic.StoreInt32(&wp.metrics.QueueSize, int32(len(wp.workQueue)))
        return nil
    case <-wp.ctx.Done():
        return fmt.Errorf("worker pool is shutting down")
    default:
        return fmt.Errorf("work queue is full")
    }
}

func (wp *WorkerPool) runWorker(worker *Worker) {
    defer wp.wg.Done()
    atomic.AddInt32(&wp.metrics.ActiveWorkers, 1)
    defer atomic.AddInt32(&wp.metrics.ActiveWorkers, -1)
    
    for {
        select {
        case entry, ok := <-worker.workQueue:
            if !ok {
                return // Channel closed
            }
            
            start := time.Now()
            err := worker.processor(entry)
            duration := time.Since(start)
            
            if err != nil {
                atomic.AddInt64(&wp.metrics.FailedJobs, 1)
                fmt.Printf("Worker %d failed to process entry: %v\n", worker.id, err)
            } else {
                atomic.AddInt64(&wp.metrics.CompletedJobs, 1)
            }
            
            // Update average processing time
            wp.updateAverageTime(duration)
            atomic.StoreInt32(&wp.metrics.QueueSize, int32(len(wp.workQueue)))
            
        case <-wp.ctx.Done():
            return
        }
    }
}

func (wp *WorkerPool) updateAverageTime(duration time.Duration) {
    // Simple moving average - in production, consider using more sophisticated metrics
    completed := atomic.LoadInt64(&wp.metrics.CompletedJobs)
    if completed > 0 {
        currentAvg := wp.metrics.AverageTime.Nanoseconds()
        newAvg := (currentAvg*int64(completed-1) + duration.Nanoseconds()) / int64(completed)
        wp.metrics.AverageTime = time.Duration(newAvg)
    }
}

func (wp *WorkerPool) Stop() {
    wp.cancel()
    close(wp.workQueue)
    wp.wg.Wait()
}

func (wp *WorkerPool) GetMetrics() WorkerMetrics {
    return WorkerMetrics{
        TotalJobs:     atomic.LoadInt64(&wp.metrics.TotalJobs),
        CompletedJobs: atomic.LoadInt64(&wp.metrics.CompletedJobs),
        FailedJobs:    atomic.LoadInt64(&wp.metrics.FailedJobs),
        AverageTime:   wp.metrics.AverageTime,
        ActiveWorkers: atomic.LoadInt32(&wp.metrics.ActiveWorkers),
        QueueSize:     atomic.LoadInt32(&wp.metrics.QueueSize),
    }
}
```

## 🎯 Step 3: Output Destinations Implementation

### 3.1 Console Output
`internal/logger/outputs/console.go`:

```go
package outputs

import (
    "fmt"
    "os"
    "sync"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
    "github.com/yourusername/shopeasy-app/internal/logger/formatters"
)

type ConsoleOutput struct {
    formatter formatters.Formatter
    colors    bool
    mu        sync.Mutex
}

func NewConsoleOutput(formatter formatters.Formatter, colors bool) *ConsoleOutput {
    return &ConsoleOutput{
        formatter: formatter,
        colors:    colors,
    }
}

func (co *ConsoleOutput) Write(entry *core.LogEntry) error {
    co.mu.Lock()
    defer co.mu.Unlock()
    
    formatted, err := co.formatter.Format(entry)
    if err != nil {
        return fmt.Errorf("failed to format log entry: %w", err)
    }
    
    var output string
    if co.colors {
        output = co.addColor(entry.Level, formatted)
    } else {
        output = formatted
    }
    
    // Write to appropriate stream based on level
    if entry.Level >= core.ErrorLevel {
        fmt.Fprintf(os.Stderr, "%s\n", output)
    } else {
        fmt.Fprintf(os.Stdout, "%s\n", output)
    }
    
    return nil
}

func (co *ConsoleOutput) addColor(level core.Level, message string) string {
    if !co.colors {
        return message
    }
    
    var colorCode string
    switch level {
    case core.DebugLevel:
        colorCode = "\033[36m" // Cyan
    case core.InfoLevel:
        colorCode = "\033[32m" // Green
    case core.WarnLevel:
        colorCode = "\033[33m" // Yellow
    case core.ErrorLevel:
        colorCode = "\033[31m" // Red
    case core.FatalLevel:
        colorCode = "\033[35m" // Magenta
    default:
        return message
    }
    
    return fmt.Sprintf("%s%s\033[0m", colorCode, message)
}

func (co *ConsoleOutput) Close() error {
    return nil // Nothing to close for console output
}
```

### 3.2 File Output
`internal/logger/outputs/file.go`:

```go
package outputs

import (
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "sync"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
    "github.com/yourusername/shopeasy-app/internal/logger/formatters"
)

type FileOutput struct {
    formatter  formatters.Formatter
    path       string
    maxSize    int64 // in bytes
    maxBackups int
    maxAge     int // in days
    compress   bool
    
    file   *os.File
    size   int64
    mu     sync.Mutex
}

func NewFileOutput(formatter formatters.Formatter, path string, maxSize int64, maxBackups int, maxAge int, compress bool) (*FileOutput, error) {
    fo := &FileOutput{
        formatter:  formatter,
        path:       path,
        maxSize:    maxSize * 1024 * 1024, // Convert MB to bytes
        maxBackups: maxBackups,
        maxAge:     maxAge,
        compress:   compress,
    }
    
    if err := fo.openFile(); err != nil {
        return nil, err
    }
    
    return fo, nil
}

func (fo *FileOutput) Write(entry *core.LogEntry) error {
    fo.mu.Lock()
    defer fo.mu.Unlock()
    
    formatted, err := fo.formatter.Format(entry)
    if err != nil {
        return fmt.Errorf("failed to format log entry: %w", err)
    }
    
    message := formatted + "\n"
    
    // Check if rotation is needed
    if fo.maxSize > 0 && fo.size+int64(len(message)) > fo.maxSize {
        if err := fo.rotate(); err != nil {
            return err
        }
    }
    
    n, err := fo.file.WriteString(message)
    if err != nil {
        return err
    }
    
    fo.size += int64(n)
    return nil
}

func (fo *FileOutput) openFile() error {
    // Ensure directory exists
    dir := filepath.Dir(fo.path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create log directory: %w", err)
    }
    
    var err error
    fo.file, err = os.OpenFile(fo.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return fmt.Errorf("failed to open log file: %w", err)
    }
    
    // Get current file size
    if stat, err := fo.file.Stat(); err == nil {
        fo.size = stat.Size()
    }
    
    return nil
}

func (fo *FileOutput) rotate() error {
    if err := fo.file.Close(); err != nil {
        return err
    }
    
    // Move current file to backup
    backupPath := fo.getBackupPath()
    if err := os.Rename(fo.path, backupPath); err != nil {
        return err
    }
    
    // Compress if enabled
    if fo.compress {
        if err := fo.compressFile(backupPath); err != nil {
            return err
        }
    }
    
    // Clean up old files
    fo.cleanup()
    
    // Open new file
    fo.size = 0
    return fo.openFile()
}

func (fo *FileOutput) getBackupPath() string {
    timestamp := time.Now().Format("2006-01-02T15-04-05.000")
    ext := filepath.Ext(fo.path)
    base := strings.TrimSuffix(fo.path, ext)
    return fmt.Sprintf("%s-%s%s", base, timestamp, ext)
}

func (fo *FileOutput) compressFile(path string) error {
    srcFile, err := os.Open(path)
    if err != nil {
        return err
    }
    defer srcFile.Close()
    
    gzPath := path + ".gz"
    gzFile, err := os.Create(gzPath)
    if err != nil {
        return err
    }
    defer gzFile.Close()
    
    gz := gzip.NewWriter(gzFile)
    defer gz.Close()
    
    if _, err := io.Copy(gz, srcFile); err != nil {
        return err
    }
    
    // Remove original file after successful compression
    return os.Remove(path)
}

func (fo *FileOutput) cleanup() {
    if fo.maxBackups <= 0 && fo.maxAge <= 0 {
        return
    }
    
    dir := filepath.Dir(fo.path)
    base := filepath.Base(fo.path)
    ext := filepath.Ext(base)
    prefix := strings.TrimSuffix(base, ext)
    
    files, err := os.ReadDir(dir)
    if err != nil {
        return
    }
    
    var backups []os.DirEntry
    cutoff := time.Now().AddDate(0, 0, -fo.maxAge)
    
    for _, file := range files {
        name := file.Name()
        if strings.HasPrefix(name, prefix+"-") && 
           (strings.HasSuffix(name, ext) || strings.HasSuffix(name, ext+".gz")) {
            
            if fo.maxAge > 0 {
                info, err := file.Info()
                if err != nil {
                    continue
                }
                if info.ModTime().Before(cutoff) {
                    os.Remove(filepath.Join(dir, name))
                    continue
                }
            }
            
            backups = append(backups, file)
        }
    }
    
    if fo.maxBackups > 0 && len(backups) > fo.maxBackups {
        // Sort by modification time (oldest first)
        sort.Slice(backups, func(i, j int) bool {
            infoI, _ := backups[i].Info()
            infoJ, _ := backups[j].Info()
            return infoI.ModTime().Before(infoJ.ModTime())
        })
        
        // Remove oldest files
        for i := 0; i < len(backups)-fo.maxBackups; i++ {
            os.Remove(filepath.Join(dir, backups[i].Name()))
        }
    }
}

func (fo *FileOutput) Close() error {
    fo.mu.Lock()
    defer fo.mu.Unlock()
    
    if fo.file != nil {
        return fo.file.Close()
    }
    return nil
}
```

### 3.3 Remote Output
`internal/logger/outputs/remote.go`:

```go
package outputs

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
    "github.com/yourusername/shopeasy-app/internal/logger/formatters"
)

type RemoteOutput struct {
    formatter     formatters.Formatter
    endpoint      string
    apiKey        string
    timeout       time.Duration
    retryAttempts int
    client        *http.Client
    buffer        []*core.LogEntry
    batchSize     int
    flushInterval time.Duration
    mu            sync.Mutex
    ctx           context.Context
    cancel        context.CancelFunc
}

type RemoteLogEntry struct {
    Timestamp string                 `json:"timestamp"`
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Fields    map[string]interface{} `json:"fields,omitempty"`
    Service   string                 `json:"service"`
    Version   string                 `json:"version"`
}

func NewRemoteOutput(formatter formatters.Formatter, endpoint, apiKey string, timeout time.Duration, retryAttempts int) *RemoteOutput {
    ctx, cancel := context.WithCancel(context.Background())
    
    ro := &RemoteOutput{
        formatter:     formatter,
        endpoint:      endpoint,
        apiKey:        apiKey,
        timeout:       timeout,
        retryAttempts: retryAttempts,
        client: &http.Client{
            Timeout: timeout,
        },
        buffer:        make([]*core.LogEntry, 0, 100),
        batchSize:     50,
        flushInterval: time.Second * 10,
        ctx:           ctx,
        cancel:        cancel,
    }
    
    // Start background flushing
    go ro.backgroundFlush()
    
    return ro
}

func (ro *RemoteOutput) Write(entry *core.LogEntry) error {
    ro.mu.Lock()
    defer ro.mu.Unlock()
    
    ro.buffer = append(ro.buffer, entry)
    
    if len(ro.buffer) >= ro.batchSize {
        return ro.flushBuffer()
    }
    
    return nil
}

func (ro *RemoteOutput) backgroundFlush() {
    ticker := time.NewTicker(ro.flushInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ro.Flush()
        case <-ro.ctx.Done():
            ro.Flush() // Final flush
            return
        }
    }
}

func (ro *RemoteOutput) Flush() error {
    ro.mu.Lock()
    defer ro.mu.Unlock()
    return ro.flushBuffer()
}

func (ro *RemoteOutput) flushBuffer() error {
    if len(ro.buffer) == 0 {
        return nil
    }
    
    entries := make([]*core.LogEntry, len(ro.buffer))
    copy(entries, ro.buffer)
    ro.buffer = ro.buffer[:0]
    
    return ro.sendEntries(entries)
}

func (ro *RemoteOutput) sendEntries(entries []*core.LogEntry) error {
    remoteEntries := make([]RemoteLogEntry, len(entries))
    
    for i, entry := range entries {
        remoteEntries[i] = RemoteLogEntry{
            Timestamp: entry.Timestamp.Format(time.RFC3339),
            Level:     entry.Level.String(),
            Message:   entry.Message,
            Fields:    entry.Fields,
            Service:   "shopeasy-api", // Could be configurable
            Version:   "1.0.0",        // Could be configurable
        }
    }
    
    payload := map[string]interface{}{
        "logs": remoteEntries,
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal log entries: %w", err)
    }
    
    return ro.sendWithRetry(jsonData)
}

func (ro *RemoteOutput) sendWithRetry(data []byte) error {
    var lastErr error
    
    for attempt := 0; attempt <= ro.retryAttempts; attempt++ {
        if attempt > 0 {
            // Exponential backoff
            backoff := time.Duration(1<<uint(attempt)) * time.Second
            time.Sleep(backoff)
        }
        
        req, err := http.NewRequestWithContext(ro.ctx, "POST", ro.endpoint, bytes.NewBuffer(data))
        if err != nil {
            lastErr = err
            continue
        }
        
        req.Header.Set("Content-Type", "application/json")
        if ro.apiKey != "" {
            req.Header.Set("Authorization", "Bearer "+ro.apiKey)
        }
        req.Header.Set("User-Agent", "ShopEasy-Logger/1.0")
        
        resp, err := ro.client.Do(req)
        if err != nil {
            lastErr = err
            continue
        }
        
        if resp.StatusCode >= 200 && resp.StatusCode < 300 {
            resp.Body.Close()
            return nil // Success
        }
        
        resp.Body.Close()
        lastErr = fmt.Errorf("remote server returned status %d", resp.StatusCode)
        
        // Don't retry for client errors (4xx)
        if resp.StatusCode >= 400 && resp.StatusCode < 500 {
            break
        }
    }
    
    return fmt.Errorf("failed to send logs after %d attempts: %w", ro.retryAttempts+1, lastErr)
}

func (ro *RemoteOutput) Close() error {
    ro.cancel()
    return ro.Flush() // Final flush
}
```

### 3.4 Output Manager
`internal/logger/outputs/manager.go`:

```go
package outputs

import (
    "fmt"
    "sync"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
    "github.com/yourusername/shopeasy-app/internal/logger/config"
    "github.com/yourusername/shopeasy-app/internal/logger/formatters"
)

type Output interface {
    Write(entry *core.LogEntry) error
    Close() error
}

type OutputManager struct {
    outputs map[string]Output
    mu      sync.RWMutex
}

func NewOutputManager() *OutputManager {
    return &OutputManager{
        outputs: make(map[string]Output),
    }
}

func (om *OutputManager) AddOutput(name string, output Output) {
    om.mu.Lock()
    defer om.mu.Unlock()
    om.outputs[name] = output
}

func (om *OutputManager) RemoveOutput(name string) {
    om.mu.Lock()
    defer om.mu.Unlock()
    
    if output, exists := om.outputs[name]; exists {
        output.Close()
        delete(om.outputs, name)
    }
}

func (om *OutputManager) WriteToOutput(name string, entry *core.LogEntry) error {
    om.mu.RLock()
    output, exists := om.outputs[name]
    om.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("output %s not found", name)
    }
    
    return output.Write(entry)
}

func (om *OutputManager) WriteToAll(entry *core.LogEntry) error {
    om.mu.RLock()
    outputs := make([]Output, 0, len(om.outputs))
    for _, output := range om.outputs {
        outputs = append(outputs, output)
    }
    om.mu.RUnlock()
    
    var errors []error
    for _, output := range outputs {
        if err := output.Write(entry); err != nil {
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("failed to write to %d outputs: %v", len(errors), errors)
    }
    
    return nil
}

func (om *OutputManager) Close() error {
    om.mu.Lock()
    defer om.mu.Unlock()
    
    var errors []error
    for name, output := range om.outputs {
        if err := output.Close(); err != nil {
            errors = append(errors, fmt.Errorf("failed to close output %s: %w", name, err))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("failed to close outputs: %v", errors)
    }
    
    return nil
}

func CreateOutputFromConfig(outputConfig config.OutputConfig) (Output, error) {
    var formatter formatters.Formatter
    
    switch outputConfig.Format {
    case "json":
        formatter = formatters.NewJSONFormatter()
    case "text":
        formatter = formatters.NewTextFormatter()
    case "pretty":
        formatter = formatters.NewPrettyFormatter(true)
    default:
        formatter = formatters.NewJSONFormatter()
    }
    
    switch outputConfig.Type {
    case "console":
        colors := true
        if val, ok := outputConfig.Config["colors"].(bool); ok {
            colors = val
        }
        return NewConsoleOutput(formatter, colors), nil
        
    case "file":
        path, ok := outputConfig.Config["path"].(string)
        if !ok {
            return nil, fmt.Errorf("file output requires 'path' configuration")
        }
        
        maxSize := int64(100) // Default 100MB
        if val, ok := outputConfig.Config["max_size"].(int); ok {
            maxSize = int64(val)
        }
        
        maxBackups := 5
        if val, ok := outputConfig.Config["max_backups"].(int); ok {
            maxBackups = val
        }
        
        maxAge := 30
        if val, ok := outputConfig.Config["max_age"].(int); ok {
            maxAge = val
        }
        
        compress := true
        if val, ok := outputConfig.Config["compress"].(bool); ok {
            compress = val
        }
        
        return NewFileOutput(formatter, path, maxSize, maxBackups, maxAge, compress)
        
    case "remote":
        endpoint, ok := outputConfig.Config["endpoint"].(string)
        if !ok {
            return nil, fmt.Errorf("remote output requires 'endpoint' configuration")
        }
        
        apiKey, _ := outputConfig.Config["api_key"].(string)
        
        timeout := 30 * time.Second
        if val, ok := outputConfig.Config["timeout"].(string); ok {
            if dur, err := time.ParseDuration(val); err == nil {
                timeout = dur
            }
        }
        
        retryAttempts := 3
        if val, ok := outputConfig.Config["retry_attempts"].(int); ok {
            retryAttempts = val
        }
        
        return NewRemoteOutput(formatter, endpoint, apiKey, timeout, retryAttempts), nil
        
    default:
        return nil, fmt.Errorf("unsupported output type: %s", outputConfig.Type)
    }
}
```

## 🎨 Step 4: Enhanced Formatters Implementation

### 4.1 Pretty Formatter with Colors
`internal/logger/formatters/pretty.go`:

```go
package formatters

import (
    "fmt"
    "strings"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
)

type PrettyFormatter struct {
    colors    bool
    showTime  bool
    showLevel bool
    showCaller bool
}

func NewPrettyFormatter(colors bool) *PrettyFormatter {
    return &PrettyFormatter{
        colors:     colors,
        showTime:   true,
        showLevel:  true,
        showCaller: true,
    }
}

func (pf *PrettyFormatter) Format(entry *core.LogEntry) (string, error) {
    var parts []string
    
    // Timestamp
    if pf.showTime {
        timeStr := entry.Timestamp.Format("15:04:05.000")
        if pf.colors {
            timeStr = pf.colorize("\033[90m", timeStr) // Dark gray
        }
        parts = append(parts, timeStr)
    }
    
    // Level
    if pf.showLevel {
        levelStr := fmt.Sprintf("[%s]", strings.ToUpper(entry.Level.String()))
        if pf.colors {
            levelStr = pf.colorizeLevel(entry.Level, levelStr)
        }
        parts = append(parts, levelStr)
    }
    
    // Caller information
    if pf.showCaller && entry.Caller != "" {
        caller := fmt.Sprintf("(%s)", entry.Caller)
        if pf.colors {
            caller = pf.colorize("\033[36m", caller) // Cyan
        }
        parts = append(parts, caller)
    }
    
    // Message
    message := entry.Message
    if pf.colors && entry.Level >= core.ErrorLevel {
        message = pf.colorize("\033[1m", message) // Bold for errors
    }
    
    result := strings.Join(parts, " ") + " " + message
    
    // Fields
    if len(entry.Fields) > 0 {
        fieldsStr := pf.formatFields(entry.Fields)
        result += " " + fieldsStr
    }
    
    return result, nil
}

func (pf *PrettyFormatter) formatFields(fields map[string]interface{}) string {
    if len(fields) == 0 {
        return ""
    }
    
    var fieldParts []string
    for key, value := range fields {
        fieldStr := fmt.Sprintf("%s=%v", key, value)
        if pf.colors {
            keyStr := pf.colorize("\033[33m", key) // Yellow
            valueStr := fmt.Sprintf("%v", value)
            fieldStr = fmt.Sprintf("%s=%s", keyStr, valueStr)
        }
        fieldParts = append(fieldParts, fieldStr)
    }
    
    fieldsStr := strings.Join(fieldParts, " ")
    if pf.colors {
        return pf.colorize("\033[90m", "{") + fieldsStr + pf.colorize("\033[90m", "}")
    }
    
    return "{" + fieldsStr + "}"
}

func (pf *PrettyFormatter) colorize(color, text string) string {
    if !pf.colors {
        return text
    }
    return color + text + "\033[0m"
}

func (pf *PrettyFormatter) colorizeLevel(level core.Level, text string) string {
    if !pf.colors {
        return text
    }
    
    var color string
    switch level {
    case core.DebugLevel:
        color = "\033[36m" // Cyan
    case core.InfoLevel:
        color = "\033[32m" // Green
    case core.WarnLevel:
        color = "\033[33m" // Yellow
    case core.ErrorLevel:
        color = "\033[31m" // Red
    case core.FatalLevel:
        color = "\033[35m\033[1m" // Bold Magenta
    default:
        return text
    }
    
    return color + text + "\033[0m"
}
```

### 4.2 Enhanced JSON Formatter
`internal/logger/formatters/json.go`:

```go
package formatters

import (
    "encoding/json"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
)

type JSONFormatter struct {
    timestampFormat string
    prettyPrint     bool
}

func NewJSONFormatter() *JSONFormatter {
    return &JSONFormatter{
        timestampFormat: time.RFC3339Nano,
        prettyPrint:     false,
    }
}

func NewPrettyJSONFormatter() *JSONFormatter {
    return &JSONFormatter{
        timestampFormat: time.RFC3339Nano,
        prettyPrint:     true,
    }
}

func (jf *JSONFormatter) Format(entry *core.LogEntry) (string, error) {
    data := make(map[string]interface{})
    
    // Core fields
    data["timestamp"] = entry.Timestamp.Format(jf.timestampFormat)
    data["level"] = entry.Level.String()
    data["message"] = entry.Message
    
    // Add caller info if available
    if entry.Caller != "" {
        data["caller"] = entry.Caller
    }
    
    // Add custom fields
    for key, value := range entry.Fields {
        // Avoid overwriting core fields
        if key != "timestamp" && key != "level" && key != "message" && key != "caller" {
            data[key] = value
        }
    }
    
    var jsonData []byte
    var err error
    
    if jf.prettyPrint {
        jsonData, err = json.MarshalIndent(data, "", "  ")
    } else {
        jsonData, err = json.Marshal(data)
    }
    
    if err != nil {
        return "", err
    }
    
    return string(jsonData), nil
}
```

### 4.3 Enhanced Text Formatter
`internal/logger/formatters/text.go`:

```go
package formatters

import (
    "fmt"
    "strings"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
)

type TextFormatter struct {
    timestampFormat string
    showCaller      bool
}

func NewTextFormatter() *TextFormatter {
    return &TextFormatter{
        timestampFormat: "2006-01-02 15:04:05.000",
        showCaller:      true,
    }
}

func (tf *TextFormatter) Format(entry *core.LogEntry) (string, error) {
    var parts []string
    
    // Timestamp
    parts = append(parts, entry.Timestamp.Format(tf.timestampFormat))
    
    // Level
    parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(entry.Level.String())))
    
    // Caller
    if tf.showCaller && entry.Caller != "" {
        parts = append(parts, fmt.Sprintf("(%s)", entry.Caller))
    }
    
    // Message
    parts = append(parts, entry.Message)
    
    result := strings.Join(parts, " ")
    
    // Fields
    if len(entry.Fields) > 0 {
        var fieldParts []string
        for key, value := range entry.Fields {
            fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", key, value))
        }
        result += " {" + strings.Join(fieldParts, " ") + "}"
    }
    
    return result, nil
}
```

## 🔍 Step 5: Specialized Loggers Implementation

### 5.1 Audit Logger
`internal/logger/specialized/audit.go`:

```go
package specialized

import (
    "context"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
)

type AuditLogger struct {
    baseLogger  *core.Logger
    outputName  string
    includeFields []string
    excludeFields []string
}

type AuditEvent struct {
    UserID    string                 `json:"user_id"`
    Action    string                 `json:"action"`
    Resource  string                 `json:"resource"`
    IPAddress string                 `json:"ip_address"`
    UserAgent string                 `json:"user_agent"`
    Success   bool                   `json:"success"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
}

func NewAuditLogger(baseLogger *core.Logger, outputName string, includeFields, excludeFields []string) *AuditLogger {
    return &AuditLogger{
        baseLogger:    baseLogger,
        outputName:    outputName,
        includeFields: includeFields,
        excludeFields: excludeFields,
    }
}

func (al *AuditLogger) LogEvent(ctx context.Context, event AuditEvent) {
    fields := make(map[string]interface{})
    
    // Add event fields
    fields["audit_type"] = "event"
    fields["user_id"] = event.UserID
    fields["action"] = event.Action
    fields["resource"] = event.Resource
    fields["ip_address"] = event.IPAddress
    fields["user_agent"] = event.UserAgent
    fields["success"] = event.Success
    
    // Add details
    for key, value := range event.Details {
        if al.shouldIncludeField(key) {
            fields[key] = value
        }
    }
    
    // Extract trace ID from context if available
    if traceID := ctx.Value("trace_id"); traceID != nil {
        fields["trace_id"] = traceID
    }
    
    message := fmt.Sprintf("Audit: %s %s by user %s", event.Action, event.Resource, event.UserID)
    if !event.Success {
        message += " (FAILED)"
    }
    
    entry := &core.LogEntry{
        Level:     core.InfoLevel,
        Message:   message,
        Fields:    fields,
        Timestamp: event.Timestamp,
    }
    
    al.baseLogger.WriteToOutput(al.outputName, entry)
}

func (al *AuditLogger) LogUserAction(ctx context.Context, userID, action, resource string, success bool, details map[string]interface{}) {
    event := AuditEvent{
        UserID:    userID,
        Action:    action,
        Resource:  resource,
        Success:   success,
        Details:   details,
        Timestamp: time.Now(),
    }
    
    // Extract IP and User-Agent from context
    if ip := ctx.Value("client_ip"); ip != nil {
        if ipStr, ok := ip.(string); ok {
            event.IPAddress = ipStr
        }
    }
    
    if ua := ctx.Value("user_agent"); ua != nil {
        if uaStr, ok := ua.(string); ok {
            event.UserAgent = uaStr
        }
    }
    
    al.LogEvent(ctx, event)
}

func (al *AuditLogger) LogSystemAction(ctx context.Context, action, resource string, success bool, details map[string]interface{}) {
    event := AuditEvent{
        UserID:    "system",
        Action:    action,
        Resource:  resource,
        Success:   success,
        Details:   details,
        Timestamp: time.Now(),
    }
    
    al.LogEvent(ctx, event)
}

func (al *AuditLogger) shouldIncludeField(field string) bool {
    // Check exclude list first
    for _, excluded := range al.excludeFields {
        if field == excluded {
            return false
        }
    }
    
    // If include list is specified, check it
    if len(al.includeFields) > 0 {
        for _, included := range al.includeFields {
            if field == included {
                return true
            }
        }
        return false
    }
    
    return true
}
```

### 5.2 Security Logger
`internal/logger/specialized/security.go`:

```go
package specialized

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
)

type SecurityLogger struct {
    baseLogger       *core.Logger
    outputName       string
    alertThreshold   int
    blocklistEnabled bool
    sensitiveFields  []string
    
    // Rate limiting and alerting
    alertCounts map[string]int
    lastReset   time.Time
    mu          sync.RWMutex
}

type SecurityEvent struct {
    EventType   string                 `json:"event_type"`
    Severity    string                 `json:"severity"`
    UserID      string                 `json:"user_id,omitempty"`
    IPAddress   string                 `json:"ip_address"`
    UserAgent   string                 `json:"user_agent,omitempty"`
    Description string                 `json:"description"`
    Details     map[string]interface{} `json:"details,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
}

const (
    SeverityLow      = "low"
    SeverityMedium   = "medium"
    SeverityHigh     = "high"
    SeverityCritical = "critical"
)

const (
    EventTypeAuthFailure     = "auth_failure"
    EventTypeSuspiciousLogin = "suspicious_login"
    EventTypeRateLimitHit    = "rate_limit_hit"
    EventTypeDataBreach      = "data_breach"
    EventTypeInjectionAttempt = "injection_attempt"
    EventTypeUnauthorizedAccess = "unauthorized_access"
    EventTypeAccountLockout  = "account_lockout"
)

func NewSecurityLogger(baseLogger *core.Logger, outputName string, alertThreshold int, blocklistEnabled bool, sensitiveFields []string) *SecurityLogger {
    return &SecurityLogger{
        baseLogger:       baseLogger,
        outputName:       outputName,
        alertThreshold:   alertThreshold,
        blocklistEnabled: blocklistEnabled,
        sensitiveFields:  sensitiveFields,
        alertCounts:      make(map[string]int),
        lastReset:        time.Now(),
    }
}

func (sl *SecurityLogger) LogEvent(ctx context.Context, event SecurityEvent) {
    // Sanitize sensitive information
    sanitizedDetails := sl.sanitizeDetails(event.Details)
    
    fields := map[string]interface{}{
        "security_event_type": event.EventType,
        "severity":           event.Severity,
        "ip_address":         event.IPAddress,
        "description":        event.Description,
    }
    
    if event.UserID != "" {
        fields["user_id"] = event.UserID
    }
    
    if event.UserAgent != "" {
        fields["user_agent"] = event.UserAgent
    }
    
    // Add sanitized details
    for key, value := range sanitizedDetails {
        fields[key] = value
    }
    
    // Extract trace ID from context
    if traceID := ctx.Value("trace_id"); traceID != nil {
        fields["trace_id"] = traceID
    }
    
    // Determine log level based on severity
    var level core.Level
    switch event.Severity {
    case SeverityLow:
        level = core.InfoLevel
    case SeverityMedium:
        level = core.WarnLevel
    case SeverityHigh:
        level = core.ErrorLevel
    case SeverityCritical:
        level = core.FatalLevel
    default:
        level = core.WarnLevel
    }
    
    message := fmt.Sprintf("Security Event [%s]: %s", strings.ToUpper(event.Severity), event.Description)
    
    entry := &core.LogEntry{
        Level:     level,
        Message:   message,
        Fields:    fields,
        Timestamp: event.Timestamp,
    }
    
    sl.baseLogger.WriteToOutput(sl.outputName, entry)
    
    // Check if we should trigger alerts
    sl.checkAlertThreshold(event)
}

func (sl *SecurityLogger) LogAuthFailure(ctx context.Context, userID, ipAddress, reason string) {
    event := SecurityEvent{
        EventType:   EventTypeAuthFailure,
        Severity:    SeverityMedium,
        UserID:      userID,
        IPAddress:   ipAddress,
        Description: fmt.Sprintf("Authentication failure for user %s: %s", userID, reason),
        Details: map[string]interface{}{
            "failure_reason": reason,
        },
        Timestamp: time.Now(),
    }
    
    sl.LogEvent(ctx, event)
}

func (sl *SecurityLogger) LogSuspiciousActivity(ctx context.Context, userID, ipAddress, activity string, details map[string]interface{}) {
    event := SecurityEvent{
        EventType:   EventTypeSuspiciousLogin,
        Severity:    SeverityHigh,
        UserID:      userID,
        IPAddress:   ipAddress,
        Description: fmt.Sprintf("Suspicious activity detected: %s", activity),
        Details:     details,
        Timestamp:   time.Now(),
    }
    
    sl.LogEvent(ctx, event)
}

func (sl *SecurityLogger) LogInjectionAttempt(ctx context.Context, ipAddress, payload, endpoint string) {
    event := SecurityEvent{
        EventType:   EventTypeInjectionAttempt,
        Severity:    SeverityCritical,
        IPAddress:   ipAddress,
        Description: "Injection attempt detected",
        Details: map[string]interface{}{
            "payload":  sl.truncatePayload(payload),
            "endpoint": endpoint,
        },
        Timestamp: time.Now(),
    }
    
    sl.LogEvent(ctx, event)
}

func (sl *SecurityLogger) LogDataBreach(ctx context.Context, userID, dataType string, recordCount int) {
    event := SecurityEvent{
        EventType:   EventTypeDataBreach,
        Severity:    SeverityCritical,
        UserID:      userID,
        Description: fmt.Sprintf("Potential data breach: %s data accessed", dataType),
        Details: map[string]interface{}{
            "data_type":    dataType,
            "record_count": recordCount,
        },
        Timestamp: time.Now(),
    }
    
    sl.LogEvent(ctx, event)
}

func (sl *SecurityLogger) sanitizeDetails(details map[string]interface{}) map[string]interface{} {
    sanitized := make(map[string]interface{})
    
    for key, value := range details {
        if sl.isSensitiveField(key) {
            sanitized[key] = "[REDACTED]"
        } else {
            sanitized[key] = value
        }
    }
    
    return sanitized
}

func (sl *SecurityLogger) isSensitiveField(field string) bool {
    fieldLower := strings.ToLower(field)
    for _, sensitive := range sl.sensitiveFields {
        if strings.ToLower(sensitive) == fieldLower {
            return true
        }
    }
    return false
}

func (sl *SecurityLogger) truncatePayload(payload string) string {
    const maxLength = 1000
    if len(payload) > maxLength {
        return payload[:maxLength] + "...[truncated]"
    }
    return payload
}

func (sl *SecurityLogger) checkAlertThreshold(event SecurityEvent) {
    if sl.alertThreshold <= 0 {
        return
    }
    
    sl.mu.Lock()
    defer sl.mu.Unlock()
    
    // Reset counts every hour
    if time.Since(sl.lastReset) > time.Hour {
        sl.alertCounts = make(map[string]int)
        sl.lastReset = time.Now()
    }
    
    key := fmt.Sprintf("%s:%s", event.EventType, event.IPAddress)
    sl.alertCounts[key]++
    
    if sl.alertCounts[key] >= sl.alertThreshold {
        // Log alert threshold reached
        alertEvent := SecurityEvent{
            EventType:   "alert_threshold_reached",
            Severity:    SeverityCritical,
            IPAddress:   event.IPAddress,
            Description: fmt.Sprintf("Alert threshold reached for %s from IP %s", event.EventType, event.IPAddress),
            Details: map[string]interface{}{
                "original_event_type": event.EventType,
                "count":              sl.alertCounts[key],
                "threshold":          sl.alertThreshold,
            },
            Timestamp: time.Now(),
        }
        
        // Use a separate context to avoid infinite recursion
        sl.LogEvent(context.Background(), alertEvent)
    }
}
```

### 5.3 Performance Logger
`internal/logger/specialized/performance.go`:

```go
package specialized

import (
    "context"
    "fmt"
    "runtime"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/logger/core"
)

type PerformanceLogger struct {
    baseLogger     *core.Logger
    outputName     string
    slowThreshold  time.Duration
    sampleRate     float64
    includeMetrics bool
    sampleCounter  int64
}

type PerformanceEvent struct {
    Operation    string                 `json:"operation"`
    Duration     time.Duration          `json:"duration_ns"`
    Success      bool                   `json:"success"`
    ErrorMessage string                 `json:"error_message,omitempty"`
    Metrics      *SystemMetrics         `json:"metrics,omitempty"`
    Details      map[string]interface{} `json:"details,omitempty"`
    Timestamp    time.Time              `json:"timestamp"`
}

type SystemMetrics struct {
    MemoryUsage    uint64  `json:"memory_usage_bytes"`
    GoroutineCount int     `json:"goroutine_count"`
    GCPauseTime    float64 `json:"gc_pause_time_ns"`
    CPUUsage       float64 `json:"cpu_usage_percent,omitempty"`
}

type Timer struct {
    logger    *PerformanceLogger
    operation string
    start     time.Time
    details   map[string]interface{}
    ctx       context.Context
}

func NewPerformanceLogger(baseLogger *core.Logger, outputName string, slowThreshold time.Duration, sampleRate float64, includeMetrics bool) *PerformanceLogger {
    return &PerformanceLogger{
        baseLogger:     baseLogger,
        outputName:     outputName,
        slowThreshold:  slowThreshold,
        sampleRate:     sampleRate,
        includeMetrics: includeMetrics,
    }
}

func (pl *PerformanceLogger) StartTimer(ctx context.Context, operation string) *Timer {
    return &Timer{
        logger:    pl,
        operation: operation,
        start:     time.Now(),
        details:   make(map[string]interface{}),
        ctx:       ctx,
    }
}

func (timer *Timer) AddDetail(key string, value interface{}) {
    timer.details[key] = value
}

func (timer *Timer) End() time.Duration {
    return timer.EndWithError(nil)
}

func (timer *Timer) EndWithError(err error) time.Duration {
    duration := time.Since(timer.start)
    
    event := PerformanceEvent{
        Operation: timer.operation,
        Duration:  duration,
        Success:   err == nil,
        Details:   timer.details,
        Timestamp: timer.start,
    }
    
    if err != nil {
        event.ErrorMessage = err.Error()
    }
    
    timer.logger.LogEvent(timer.ctx, event)
    return duration
}

func (pl *PerformanceLogger) LogEvent(ctx context.Context, event PerformanceEvent) {
    // Apply sampling
    pl.sampleCounter++
    if pl.sampleRate < 1.0 {
        if float64(pl.sampleCounter%100)/100.0 > pl.sampleRate {
            return
        }
    }
    
    // Only log if it's slow or failed
    if event.Success && event.Duration < pl.slowThreshold {
        return
    }
    
    fields := map[string]interface{}{
        "performance_event": true,
        "operation":         event.Operation,
        "duration_ms":       float64(event.Duration.Nanoseconds()) / 1000000.0,
        "duration_ns":       event.Duration.Nanoseconds(),
        "success":           event.Success,
    }
    
    if event.ErrorMessage != "" {
        fields["error"] = event.ErrorMessage
    }
    
    // Add details
    for key, value := range event.Details {
        fields[key] = value
    }
    
    // Add system metrics if enabled
    if pl.includeMetrics {
        metrics := pl.collectSystemMetrics()
        if metrics != nil {
            fields["memory_usage_mb"] = float64(metrics.MemoryUsage) / 1024 / 1024
            fields["goroutine_count"] = metrics.GoroutineCount
            fields["gc_pause_time_ms"] = metrics.GCPauseTime / 1000000.0
            if metrics.CPUUsage > 0 {
                fields["cpu_usage_percent"] = metrics.CPUUsage
            }
        }
    }
    
    // Extract trace ID from context
    if traceID := ctx.Value("trace_id"); traceID != nil {
        fields["trace_id"] = traceID
    }
    
    // Determine log level
    var level core.Level
    if !event.Success {
        level = core.ErrorLevel
    } else if event.Duration > pl.slowThreshold*2 {
        level = core.WarnLevel
    } else {
        level = core.InfoLevel
    }
    
    var message string
    if event.Success {
        message = fmt.Sprintf("Slow operation: %s took %v", event.Operation, event.Duration)
    } else {
        message = fmt.Sprintf("Failed operation: %s took %v - %s", event.Operation, event.Duration, event.ErrorMessage)
    }
    
    entry := &core.LogEntry{
        Level:     level,
        Message:   message,
        Fields:    fields,
        Timestamp: event.Timestamp,
    }
    
    pl.baseLogger.WriteToOutput(pl.outputName, entry)
}

func (pl *PerformanceLogger) LogSlowQuery(ctx context.Context, query string, duration time.Duration, rowsAffected int64) {
    event := PerformanceEvent{
        Operation: "database_query",
        Duration:  duration,
        Success:   true,
        Details: map[string]interface{}{
            "query":         query,
            "rows_affected": rowsAffected,
        },
        Timestamp: time.Now().Add(-duration),
    }