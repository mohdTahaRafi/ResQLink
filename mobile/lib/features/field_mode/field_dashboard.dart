import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import '../../core/connectivity.dart';
import '../../data/remote/storage_service.dart';
import '../../main.dart'; // for apiClientProvider

/// NGO Worker "Field Mode" — High-contrast UI for data entry with offline support.
class FieldDashboard extends ConsumerStatefulWidget {
  const FieldDashboard({super.key});

  @override
  ConsumerState<FieldDashboard> createState() => _FieldDashboardState();
}

class _FieldDashboardState extends ConsumerState<FieldDashboard> {
  final _descriptionController = TextEditingController();
  final _connectivityService = ConnectivityService();
  final _storageService = StorageService();
  String? _selectedMediaPath;
  XFile? _selectedMediaFile;
  bool _isSubmitting = false;
  int _pendingSync = 0;

  // Hardcoded Lucknow coordinates for demo — replace with geolocator in production
  static const double _defaultLat = 26.8467;
  static const double _defaultLng = 80.9462;

  @override
  void initState() {
    super.initState();
    _pendingSync = _connectivityService.pendingCount;
    _connectivityService.startMonitoring(onReconnect: _syncPending);
  }

  @override
  void dispose() {
    _descriptionController.dispose();
    _connectivityService.dispose();
    super.dispose();
  }

  Future<void> _pickImage() async {
    final picker = ImagePicker();
    // Use gallery on web/desktop, camera on mobile
    final photo = await picker.pickImage(source: ImageSource.gallery, maxWidth: 1280);
    if (photo != null) {
      setState(() {
        _selectedMediaPath = photo.path;
        _selectedMediaFile = photo;
      });
    }
  }

  Future<void> _submitReport() async {
    if (_descriptionController.text.isEmpty && _selectedMediaPath == null) return;

    setState(() => _isSubmitting = true);

    final reportData = {
      'raw_text': _descriptionController.text,
      'media_type': _selectedMediaPath != null ? 'image' : 'text',
      'media_path': _selectedMediaPath ?? '',
      'latitude': _defaultLat,
      'longitude': _defaultLng,
      'timestamp': DateTime.now().toIso8601String(),
    };

    if (_connectivityService.isOnline) {
      try {
        final api = ref.read(apiClientProvider);

        // Upload image to Firebase Storage if one was selected
        String mediaUrl = '';
        if (_selectedMediaFile != null) {
          _showSuccess('Uploading image...');
          final url = await _storageService.uploadReportImage(_selectedMediaFile!);
          if (url != null) {
            mediaUrl = url;
          } else {
            _showSuccess('Image upload failed, submitting text only');
          }
        }

        final response = await api.createReport(
          rawText: _descriptionController.text,
          mediaType: _selectedMediaFile != null ? 'image' : 'text',
          mediaUrl: mediaUrl,
          latitude: _defaultLat,
          longitude: _defaultLng,
        );
        _showSuccess('Report submitted! ID: ${response['id']}');
      } catch (e) {
        // If API call fails, queue it offline
        debugPrint('API ERROR: $e');
        await _connectivityService.queueReport(reportData);
        setState(() => _pendingSync = _connectivityService.pendingCount);
        _showSuccess('API error — saved offline for retry');
      }
    } else {
      await _connectivityService.queueReport(reportData);
      setState(() => _pendingSync = _connectivityService.pendingCount);
      _showSuccess('Saved offline — will sync when connected');
    }

    _descriptionController.clear();
    setState(() {
      _selectedMediaPath = null;
      _selectedMediaFile = null;
      _isSubmitting = false;
    });
  }

  Future<void> _syncPending() async {
    final api = ref.read(apiClientProvider);
    final synced = await _connectivityService.syncPendingReports(
      submitFn: (data) async {
        try {
          await api.createReport(
            rawText: (data['raw_text'] as String?) ?? '',
            mediaType: (data['media_type'] as String?) ?? 'text',
            mediaUrl: '',
            latitude: (data['latitude'] as num?)?.toDouble() ?? _defaultLat,
            longitude: (data['longitude'] as num?)?.toDouble() ?? _defaultLng,
          );
          return true;
        } catch (_) {
          return false;
        }
      },
    );
    if (mounted) {
      setState(() => _pendingSync = _connectivityService.pendingCount);
      if (synced > 0) _showSuccess('Synced $synced pending reports');
    }
  }

  void _showSuccess(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), behavior: SnackBarBehavior.floating),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Field Mode'),
        actions: [
          if (_pendingSync > 0)
            Badge(
              label: Text('$_pendingSync'),
              child: IconButton(
                icon: const Icon(Icons.sync),
                onPressed: _syncPending,
              ),
            ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await FirebaseAuth.instance.signOut();
              if (mounted) context.go('/login');
            },
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Connectivity indicator
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: _connectivityService.isOnline
                    ? Colors.green.withOpacity(0.1)
                    : Colors.orange.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    _connectivityService.isOnline ? Icons.wifi : Icons.wifi_off,
                    size: 16,
                    color: _connectivityService.isOnline ? Colors.green : Colors.orange,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    _connectivityService.isOnline ? 'Online' : 'Offline Mode',
                    style: theme.textTheme.labelMedium,
                  ),
                ],
              ),
            ),
            const SizedBox(height: 24),

            Text('New Report', style: theme.textTheme.titleLarge),
            const SizedBox(height: 16),

            // Photo capture
            GestureDetector(
              onTap: _pickImage,
              child: Container(
                height: 180,
                width: double.infinity,
                decoration: BoxDecoration(
                  color: theme.colorScheme.surfaceContainerHighest,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(
                    color: theme.colorScheme.outlineVariant,
                    width: 2,
                    strokeAlign: BorderSide.strokeAlignInside,
                  ),
                ),
                child: _selectedMediaPath != null
                    ? ClipRRect(
                        borderRadius: BorderRadius.circular(14),
                        child: Image.network(
                          _selectedMediaPath!,
                          fit: BoxFit.cover,
                          errorBuilder: (_, __, ___) => const Center(
                            child: Icon(Icons.broken_image, size: 48),
                          ),
                        ),
                      )
                    : Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.camera_alt_rounded,
                              size: 48, color: theme.colorScheme.primary),
                          const SizedBox(height: 8),
                          Text('Tap to capture photo',
                              style: theme.textTheme.bodyMedium),
                        ],
                      ),
              ),
            ),
            const SizedBox(height: 16),

            // Description
            TextFormField(
              controller: _descriptionController,
              maxLines: 4,
              decoration: const InputDecoration(
                labelText: 'Describe the issue',
                hintText: 'Enter details in Hindi or English...',
                alignLabelWithHint: true,
              ),
            ),
            const Spacer(),

            // Submit
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: _isSubmitting ? null : _submitReport,
                icon: _isSubmitting
                    ? const SizedBox(
                        height: 18, width: 18,
                        child: CircularProgressIndicator(strokeWidth: 2))
                    : const Icon(Icons.send_rounded),
                label: Text(_isSubmitting ? 'Submitting...' : 'Submit Report'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
