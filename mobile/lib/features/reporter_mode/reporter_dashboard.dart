import '../../data/local/hive_store.dart';
import 'dart:convert';
import 'dart:typed_data';
import 'dart:async';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import '../../main.dart';
// Web geolocation
import 'dart:js_interop' as js;
import 'package:web/web.dart' as web;

/// General User "Reporter" Dashboard — Submit issues with full details.
class ReporterDashboard extends ConsumerStatefulWidget {
  const ReporterDashboard({super.key});

  @override
  ConsumerState<ReporterDashboard> createState() => _ReporterDashboardState();
}

class _ReporterDashboardState extends ConsumerState<ReporterDashboard> {
  final _descriptionController = TextEditingController();
  final _locationController = TextEditingController();
  final _volunteerCountController = TextEditingController(text: '1');
  final List<Uint8List> _selectedImages = [];
  bool _isSubmitting = false;
  bool _isGettingLocation = false;

  String _issueType = 'civic_issue';
  String _urgency = 'normal';

  // User's GPS coordinates (null until fetched)
  double? _latitude;
  double? _longitude;

  static const _issueTypes = {
    'medical_emergency': 'Medical Emergency',
    'legal_aid': 'Legal Aid',
    'civic_issue': 'Civic Issue',
    'disaster_relief': 'Disaster Relief',
  };

  static const _urgencyLevels = {
    'normal': 'Normal',
    'urgent': 'Urgent',
    'critical': 'Critical',
  };

  @override
  void dispose() {
    _descriptionController.dispose();
    _locationController.dispose();
    _volunteerCountController.dispose();
    super.dispose();
  }

  Future<void> _pickImages() async {
    final picker = ImagePicker();
    final photos = await picker.pickMultiImage(maxWidth: 800, imageQuality: 50);
    for (final photo in photos) {
      final bytes = await photo.readAsBytes();
      setState(() => _selectedImages.add(bytes));
    }
  }

  Future<void> _getLocation() async {
    setState(() => _isGettingLocation = true);
    try {
      if (kIsWeb) {
        final nav = web.window.navigator;
        final completer = Completer<void>();
        nav.geolocation.getCurrentPosition(
          ((web.GeolocationPosition pos) {
            setState(() {
              _latitude = pos.coords.latitude.toDouble();
              _longitude = pos.coords.longitude.toDouble();
              _isGettingLocation = false;
            });
            _showSnackBar('Location captured!');
            completer.complete();
          }).toJS,
          ((web.GeolocationPositionError err) {
            setState(() => _isGettingLocation = false);
            _showSnackBar('Location denied. Using default.');
            completer.complete();
          }).toJS,
        );
        await completer.future.timeout(const Duration(seconds: 10), onTimeout: () {
          setState(() => _isGettingLocation = false);
          _showSnackBar('Location timeout.');
        });
      }
    } catch (e) {
      setState(() => _isGettingLocation = false);
      _showSnackBar('Could not get location');
    }
  }

  Future<void> _submitReport() async {
    if (_descriptionController.text.isEmpty) {
      _showSnackBar('Please describe the issue');
      return;
    }
    if (_locationController.text.isEmpty) {
      _showSnackBar('Location is required');
      return;
    }

    setState(() => _isSubmitting = true);

    // Encode first image as base64 (API supports one for now)
    String mediaBase64 = '';
    if (_selectedImages.isNotEmpty) {
      mediaBase64 = 'data:image/jpeg;base64,${base64Encode(_selectedImages.first)}';
    }

    try {
      final api = ref.read(apiClientProvider);
      final response = await api.createReport(
        rawText: _descriptionController.text,
        mediaType: _selectedImages.isNotEmpty ? 'image' : 'text',
        mediaUrl: mediaBase64,
        latitude: _latitude ?? 26.8467,
        longitude: _longitude ?? 80.9462,
        issueType: _issueType,
        userUrgency: _urgency,
        requiredVolunteers: int.tryParse(_volunteerCountController.text) ?? 1,
        location: _locationController.text,
      );
      _showSnackBar('Issue reported! ID: ${response['id']}');
      _descriptionController.clear();
      _locationController.clear();
      _volunteerCountController.text = '1';
      setState(() {
        _selectedImages.clear();
        _issueType = 'civic_issue';
        _urgency = 'normal';
      });
    } catch (e) {
      debugPrint('API ERROR: $e');
      _showSnackBar('Error submitting report. Please try again.');
    }

    setState(() => _isSubmitting = false);
  }

  void _showSnackBar(String msg) {
    if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Report an Issue'),
        actions: [
          IconButton(
            icon: const Icon(Icons.auto_awesome),
            tooltip: 'AI Hub',
            onPressed: () => context.push('/ai-hub'),
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await HiveStore.clearAll();
              await FirebaseAuth.instance.signOut();
              if (mounted) context.go('/role-select');
            },
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // --- Photo Upload ---
            Text('Photos', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600)),
            const SizedBox(height: 8),
            GestureDetector(
              onTap: _pickImages,
              child: Container(
                height: _selectedImages.isEmpty ? 120 : null,
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surfaceContainerHighest,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: theme.colorScheme.outlineVariant, width: 2),
                ),
                child: _selectedImages.isEmpty
                    ? Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.add_a_photo_rounded, size: 36, color: theme.colorScheme.primary),
                          const SizedBox(height: 8),
                          Text('Tap to add photos', style: theme.textTheme.bodySmall),
                        ],
                      )
                    : Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          ..._selectedImages.asMap().entries.map((e) => Stack(
                                children: [
                                  ClipRRect(
                                    borderRadius: BorderRadius.circular(8),
                                    child: Image.memory(e.value, width: 80, height: 80, fit: BoxFit.cover),
                                  ),
                                  Positioned(
                                    top: 0, right: 0,
                                    child: GestureDetector(
                                      onTap: () => setState(() => _selectedImages.removeAt(e.key)),
                                      child: Container(
                                        decoration: const BoxDecoration(color: Colors.red, shape: BoxShape.circle),
                                        child: const Icon(Icons.close, size: 16, color: Colors.white),
                                      ),
                                    ),
                                  ),
                                ],
                              )),
                          GestureDetector(
                            onTap: _pickImages,
                            child: Container(
                              width: 80, height: 80,
                              decoration: BoxDecoration(
                                border: Border.all(color: theme.colorScheme.primary, width: 2),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: Icon(Icons.add, color: theme.colorScheme.primary),
                            ),
                          ),
                        ],
                      ),
              ),
            ),
            const SizedBox(height: 20),

            // --- Description ---
            Text('Description *', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600)),
            const SizedBox(height: 8),
            TextField(
              controller: _descriptionController,
              maxLines: 4,
              decoration: const InputDecoration(hintText: 'Describe the issue in detail...', border: OutlineInputBorder()),
            ),
            const SizedBox(height: 20),

            // --- Issue Type ---
            Text('Issue Type *', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600)),
            const SizedBox(height: 8),
            DropdownButtonFormField<String>(
              value: _issueType,
              decoration: const InputDecoration(border: OutlineInputBorder()),
              items: _issueTypes.entries.map((e) => DropdownMenuItem(value: e.key, child: Text(e.value))).toList(),
              onChanged: (v) => setState(() => _issueType = v!),
            ),
            const SizedBox(height: 20),

            // --- Location ---
            Text('Location *', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600)),
            const SizedBox(height: 8),
            TextField(
              controller: _locationController,
              decoration: const InputDecoration(
                hintText: 'Enter address or area name',
                border: OutlineInputBorder(),
                prefixIcon: Icon(Icons.location_on),
              ),
            ),
            const SizedBox(height: 12),

            // --- GPS Coordinates ---
            Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _isGettingLocation ? null : _getLocation,
                    icon: _isGettingLocation
                        ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Icon(Icons.my_location, size: 18),
                    label: Text(_latitude != null
                        ? 'Lat: ${_latitude!.toStringAsFixed(4)}, Lng: ${_longitude!.toStringAsFixed(4)}'
                        : 'Get My Location'),
                  ),
                ),
              ],
            ),
            if (_latitude != null)
              Padding(
                padding: const EdgeInsets.only(top: 4),
                child: Text('✓ GPS coordinates captured', style: theme.textTheme.labelSmall?.copyWith(color: Colors.green)),
              ),
            const SizedBox(height: 20),

            // --- Urgency ---
            Text('Urgency (optional)', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600)),
            const SizedBox(height: 8),
            Row(
              children: _urgencyLevels.entries.map((e) {
                final isSelected = e.key == _urgency;
                final color = e.key == 'critical' ? Colors.red : (e.key == 'urgent' ? Colors.orange : Colors.green);
                return Expanded(
                  child: Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 4),
                    child: ChoiceChip(
                      label: Text(e.value),
                      selected: isSelected,
                      selectedColor: color.withOpacity(0.2),
                      onSelected: (_) => setState(() => _urgency = e.key),
                    ),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 20),

            // --- Required Volunteers ---
            Text('Required Volunteers', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600)),
            const SizedBox(height: 8),
            TextField(
              controller: _volunteerCountController,
              keyboardType: TextInputType.number,
              decoration: const InputDecoration(
                hintText: 'How many volunteers needed?',
                border: OutlineInputBorder(),
                prefixIcon: Icon(Icons.people),
              ),
            ),
            const SizedBox(height: 32),

            // --- Submit ---
            SizedBox(
              width: double.infinity,
              height: 52,
              child: FilledButton.icon(
                onPressed: _isSubmitting ? null : _submitReport,
                icon: _isSubmitting ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)) : const Icon(Icons.send),
                label: Text(_isSubmitting ? 'Submitting...' : 'Submit Report'),
              ),
            ),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }
}
