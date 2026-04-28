import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import '../../main.dart';

/// AI Hub — showcases all 10 Gemini-powered AI features.
class AIHubScreen extends ConsumerStatefulWidget {
  const AIHubScreen({super.key});

  @override
  ConsumerState<AIHubScreen> createState() => _AIHubScreenState();
}

class _AIHubScreenState extends ConsumerState<AIHubScreen> {
  int _selectedFeature = 0;

  final _features = const [
    _AIFeature('AI Chat', Icons.smart_toy_outlined, Color(0xFF6C63FF)),
    _AIFeature('Image Analysis', Icons.image_search, Color(0xFF00BFA5)),
    _AIFeature('Verification', Icons.verified_user, Color(0xFF795548)),
    _AIFeature('Sentiment', Icons.mood, Color(0xFFFF6F61)),
    _AIFeature('Translation', Icons.translate, Color(0xFF2196F3)),
    _AIFeature('Duplicates', Icons.content_copy, Color(0xFFFF9800)),
    _AIFeature('Action Plan', Icons.task_alt, Color(0xFF607D8B)),
    _AIFeature('OCR', Icons.document_scanner, Color(0xFF9C27B0)),
    _AIFeature('Progress', Icons.analytics, Color(0xFF4CAF50)),
    _AIFeature('Skills', Icons.school, Color(0xFFE91E63)),
  ];

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(6),
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  colors: [Color(0xFF6C63FF), Color(0xFF00BFA5)],
                ),
                borderRadius: BorderRadius.circular(8),
              ),
              child:
                  const Icon(Icons.auto_awesome, color: Colors.white, size: 18),
            ),
            const SizedBox(width: 10),
            const Text('AI Hub'),
          ],
        ),
        centerTitle: true,
      ),
      body: Column(
        children: [
          // Feature selector chips
          Container(
            height: 50,
            margin: const EdgeInsets.symmetric(vertical: 8),
            child: ListView.builder(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 12),
              itemCount: _features.length,
              itemBuilder: (ctx, i) {
                final f = _features[i];
                final selected = i == _selectedFeature;
                return Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  child: ChoiceChip(
                    avatar: Icon(f.icon,
                        size: 18, color: selected ? Colors.white : f.color),
                    label: Text(f.label,
                        style: TextStyle(
                          color: selected
                              ? Colors.white
                              : theme.colorScheme.onSurface,
                          fontWeight:
                              selected ? FontWeight.w600 : FontWeight.w400,
                          fontSize: 12,
                        )),
                    selected: selected,
                    selectedColor: f.color,
                    onSelected: (_) => setState(() => _selectedFeature = i),
                  ),
                );
              },
            ),
          ),
          const Divider(height: 1),
          // Feature content
          Expanded(
            child: IndexedStack(
              index: _selectedFeature,
              children: [
                _ChatTab(),
                _ImageAnalysisTab(),
                _VerificationTab(),
                _SentimentTab(),
                _TranslationTab(),
                _DuplicateTab(),
                _ActionPlanTab(),
                _OCRTab(),
                _ProgressReportTab(),
                _SkillsTab(),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _AIFeature {
  final String label;
  final IconData icon;
  final Color color;
  const _AIFeature(this.label, this.icon, this.color);
}

String _imagePayload(Uint8List bytes, String? name) {
  final lower = (name ?? '').toLowerCase();
  final mime = lower.endsWith('.png')
      ? 'image/png'
      : lower.endsWith('.webp')
          ? 'image/webp'
          : 'image/jpeg';
  return 'data:$mime;base64,${base64Encode(bytes)}';
}

// ══════════════════════════════════════════════════
// TAB 1: AI CHATBOT
// ══════════════════════════════════════════════════

class _ChatTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_ChatTab> createState() => _ChatTabState();
}

class _ChatTabState extends ConsumerState<_ChatTab> {
  final _ctrl = TextEditingController();
  final _messages = <_ChatMsg>[];
  bool _loading = false;

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  Future<void> _send() async {
    final text = _ctrl.text.trim();
    if (text.isEmpty) return;
    setState(() {
      _messages.add(_ChatMsg(text, true));
      _loading = true;
    });
    _ctrl.clear();
    try {
      final api = ref.read(apiClientProvider);
      final response = await api.chatWithAI(text);
      setState(() => _messages.add(_ChatMsg(response, false)));
    } catch (e) {
      setState(() => _messages.add(_ChatMsg('Error: $e', false)));
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      children: [
        if (_messages.isEmpty)
          Expanded(
            child: Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.smart_toy,
                      size: 64,
                      color: theme.colorScheme.primary.withValues(alpha: 0.3)),
                  const SizedBox(height: 16),
                  Text('SAMAJ AI Assistant',
                      style: theme.textTheme.titleMedium?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      )),
                  const SizedBox(height: 8),
                  Text(
                    'Ask me anything about your tasks, safety protocols,\nor how to handle specific situations.',
                    textAlign: TextAlign.center,
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: theme.colorScheme.onSurfaceVariant),
                  ),
                ],
              ),
            ),
          )
        else
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: _messages.length,
              itemBuilder: (ctx, i) {
                final msg = _messages[i];
                return Align(
                  alignment:
                      msg.isUser ? Alignment.centerRight : Alignment.centerLeft,
                  child: Container(
                    margin: const EdgeInsets.only(bottom: 8),
                    padding: const EdgeInsets.symmetric(
                        horizontal: 14, vertical: 10),
                    constraints: BoxConstraints(
                        maxWidth: MediaQuery.of(context).size.width * 0.75),
                    decoration: BoxDecoration(
                      color: msg.isUser
                          ? theme.colorScheme.primary
                          : theme.colorScheme.surfaceContainerHighest,
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Text(msg.text,
                        style: TextStyle(
                          color: msg.isUser
                              ? theme.colorScheme.onPrimary
                              : theme.colorScheme.onSurface,
                        )),
                  ),
                );
              },
            ),
          ),
        if (_loading) const LinearProgressIndicator(),
        Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _ctrl,
                  decoration: InputDecoration(
                    hintText: 'Ask SAMAJ AI...',
                    border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(24)),
                    contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16, vertical: 12),
                  ),
                  onSubmitted: (_) => _send(),
                ),
              ),
              const SizedBox(width: 8),
              FloatingActionButton.small(
                onPressed: _loading ? null : _send,
                child: const Icon(Icons.send),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _ChatMsg {
  final String text;
  final bool isUser;
  _ChatMsg(this.text, this.isUser);
}

// ══════════════════════════════════════════════════
// TAB 2: IMAGE ANALYSIS
// ══════════════════════════════════════════════════

class _ImageAnalysisTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_ImageAnalysisTab> createState() => _ImageAnalysisTabState();
}

class _ImageAnalysisTabState extends ConsumerState<_ImageAnalysisTab> {
  Map<String, dynamic>? _result;
  bool _loading = false;
  Uint8List? _imageBytes;
  String? _imageName;
  String? _error;

  Future<void> _pickImage() async {
    final image = await ImagePicker().pickImage(
      source: ImageSource.gallery,
      maxWidth: 1400,
      imageQuality: 80,
    );
    if (image == null) return;
    final bytes = await image.readAsBytes();
    setState(() {
      _imageBytes = bytes;
      _imageName = image.name;
      _result = null;
      _error = null;
    });
  }

  Future<void> _analyze() async {
    if (_imageBytes == null) {
      await _pickImage();
      if (_imageBytes == null) return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final api = ref.read(apiClientProvider);
      final payload = _imagePayload(_imageBytes!, _imageName);
      _result = await api.analyzeImage(payload);
      setState(() {});
    } catch (e) {
      setState(() => _error = 'Analysis failed: $e');
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _FeatureHeader(
            icon: Icons.image_search,
            color: const Color(0xFF00BFA5),
            title: 'Multimodal Image Analysis',
            subtitle:
                'Upload a photo and Gemini will detect the issue type, severity, and key objects',
          ),
          const SizedBox(height: 16),
          OutlinedButton.icon(
            onPressed: _loading ? null : _pickImage,
            icon: const Icon(Icons.upload_file),
            label: Text(_imageName == null ? 'Choose Image' : 'Change Image'),
          ),
          if (_imageBytes != null) ...[
            const SizedBox(height: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(12),
              child: Image.memory(_imageBytes!, height: 180, fit: BoxFit.cover),
            ),
            const SizedBox(height: 8),
            Text(_imageName ?? 'Selected image',
                textAlign: TextAlign.center,
                style: theme.textTheme.bodySmall),
          ],
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: _loading || _imageBytes == null ? null : _analyze,
            icon: _loading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.auto_awesome),
            label: Text(_loading ? 'Analyzing...' : 'Analyze Image'),
          ),
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(_error!, style: TextStyle(color: theme.colorScheme.error)),
          ],
          if (_result != null) ...[
            const SizedBox(height: 16),
            _ResultCard(
                title: 'Issue Type',
                value: _result!['issue_type']?.toString() ?? 'N/A'),
            _ResultCard(
                title: 'Description',
                value: _result!['description']?.toString() ?? 'N/A'),
            _ResultCard(title: 'Severity', value: '${_result!['severity']}/10'),
            _ResultCard(
                title: 'Category',
                value: _result!['suggested_category']?.toString() ?? 'N/A'),
          ],
        ],
      ),
    );
  }
}

// ══════════════════════════════════════════════════
// TAB 3: REPORT VERIFICATION
// ══════════════════════════════════════════════════

class _VerificationTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_VerificationTab> createState() => _VerificationTabState();
}

class _VerificationTabState extends ConsumerState<_VerificationTab> {
  final _ctrl = TextEditingController();
  Uint8List? _imageBytes;
  String? _imageName;
  Map<String, dynamic>? _result;
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  Future<void> _pickImage() async {
    final image = await ImagePicker().pickImage(
      source: ImageSource.gallery,
      maxWidth: 1400,
      imageQuality: 80,
    );
    if (image == null) return;
    final bytes = await image.readAsBytes();
    setState(() {
      _imageBytes = bytes;
      _imageName = image.name;
      _result = null;
      _error = null;
    });
  }

  Future<void> _verify() async {
    if (_ctrl.text.trim().isEmpty) return;
    if (_imageBytes == null) {
      await _pickImage();
      if (_imageBytes == null) return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final api = ref.read(apiClientProvider);
      _result = await api.verifyReport(
        _imagePayload(_imageBytes!, _imageName),
        _ctrl.text.trim(),
      );
      setState(() {});
    } catch (e) {
      setState(() => _error = 'Verification failed: $e');
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _FeatureHeader(
            icon: Icons.verified_user,
            color: const Color(0xFF795548),
            title: 'AI Report Verification',
            subtitle: 'Compare a report photo with the written description',
          ),
          const SizedBox(height: 16),
          OutlinedButton.icon(
            onPressed: _loading ? null : _pickImage,
            icon: const Icon(Icons.upload_file),
            label: Text(_imageName == null ? 'Choose Image' : 'Change Image'),
          ),
          if (_imageBytes != null) ...[
            const SizedBox(height: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(12),
              child: Image.memory(_imageBytes!, height: 160, fit: BoxFit.cover),
            ),
          ],
          const SizedBox(height: 12),
          TextField(
            controller: _ctrl,
            maxLines: 4,
            decoration: InputDecoration(
              hintText: 'Describe what the report claims...',
              border:
                  OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            ),
          ),
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: _loading || _imageBytes == null ? null : _verify,
            icon: _loading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.fact_check),
            label: Text(_loading ? 'Verifying...' : 'Verify Report'),
            style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFF795548)),
          ),
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(_error!, style: TextStyle(color: theme.colorScheme.error)),
          ],
          if (_result != null) ...[
            const SizedBox(height: 16),
            _ResultCard(
                title: 'Consistent',
                value: _result!['is_consistent'] == true ? 'Yes' : 'No'),
            _ResultCard(
                title: 'Confidence',
                value:
                    '${((_result!['confidence'] ?? 0) * 100).toStringAsFixed(0)}%'),
            _ResultCard(
                title: 'Summary',
                value:
                    _result!['verification_summary']?.toString() ?? 'N/A'),
            if (_result!['discrepancies'] != null)
              _ResultCard(
                  title: 'Discrepancies',
                  value: (_result!['discrepancies'] as List).join(', ')),
          ],
        ],
      ),
    );
  }
}

// ══════════════════════════════════════════════════
// TAB 4: SENTIMENT ANALYSIS
// ══════════════════════════════════════════════════

class _SentimentTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_SentimentTab> createState() => _SentimentTabState();
}

class _SentimentTabState extends ConsumerState<_SentimentTab> {
  final _ctrl = TextEditingController();
  Map<String, dynamic>? _result;
  bool _loading = false;

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  Future<void> _analyze() async {
    if (_ctrl.text.trim().isEmpty) return;
    setState(() => _loading = true);
    try {
      final api = ref.read(apiClientProvider);
      _result = await api.analyzeSentiment(_ctrl.text.trim());
      setState(() {});
    } catch (e) {
      setState(() => _result = {'overall_sentiment': 'Error: $e'});
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _FeatureHeader(
            icon: Icons.mood,
            color: const Color(0xFFFF6F61),
            title: 'Sentiment & Emotion Analysis',
            subtitle:
                'Detect emotional tone, urgency signals, and distress levels',
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _ctrl,
            maxLines: 4,
            decoration: InputDecoration(
              hintText: 'Paste or type a report text to analyze...',
              border:
                  OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            ),
          ),
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: _loading ? null : _analyze,
            icon: _loading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.psychology),
            label: Text(_loading ? 'Analyzing...' : 'Analyze Sentiment'),
            style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFFFF6F61)),
          ),
          if (_result != null) ...[
            const SizedBox(height: 16),
            _ResultCard(
                title: 'Sentiment',
                value: _result!['overall_sentiment']?.toString() ?? 'N/A'),
            _ResultCard(
                title: 'Urgency Boost', value: '+${_result!['urgency_boost']}'),
            _ResultCard(
                title: 'Emotional Intensity',
                value:
                    '${((_result!['emotional_intensity'] ?? 0) * 100).toStringAsFixed(0)}%'),
            if (_result!['emotions_detected'] != null)
              _ResultCard(
                  title: 'Emotions',
                  value: (_result!['emotions_detected'] as List).join(', ')),
            if (_result!['recommendation'] != null)
              _ResultCard(
                  title: 'Recommendation',
                  value: _result!['recommendation'].toString()),
          ],
        ],
      ),
    );
  }
}

// ══════════════════════════════════════════════════
// TAB 4: TRANSLATION
// ══════════════════════════════════════════════════

class _TranslationTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_TranslationTab> createState() => _TranslationTabState();
}

class _TranslationTabState extends ConsumerState<_TranslationTab> {
  final _ctrl = TextEditingController();
  String _targetLang = 'Hindi';
  Map<String, dynamic>? _result;
  bool _loading = false;

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  Future<void> _translate() async {
    if (_ctrl.text.trim().isEmpty) return;
    setState(() => _loading = true);
    try {
      final api = ref.read(apiClientProvider);
      _result = await api.translate(_ctrl.text.trim(), targetLang: _targetLang);
      setState(() {});
    } catch (e) {
      setState(() => _result = {'translated_text': 'Error: $e'});
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _FeatureHeader(
            icon: Icons.translate,
            color: const Color(0xFF2196F3),
            title: 'Real-Time Translation',
            subtitle:
                'Translate between Hindi, English, and regional languages',
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _ctrl,
            maxLines: 3,
            decoration: InputDecoration(
              hintText: 'Enter text to translate...',
              border:
                  OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            ),
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              const Text('Target: '),
              const SizedBox(width: 8),
              SegmentedButton<String>(
                segments: const [
                  ButtonSegment(value: 'Hindi', label: Text('Hindi')),
                  ButtonSegment(value: 'English', label: Text('English')),
                  ButtonSegment(value: 'Tamil', label: Text('Tamil')),
                ],
                selected: {_targetLang},
                onSelectionChanged: (s) =>
                    setState(() => _targetLang = s.first),
              ),
            ],
          ),
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: _loading ? null : _translate,
            icon: _loading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.translate),
            label: Text(_loading ? 'Translating...' : 'Translate'),
            style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFF2196F3)),
          ),
          if (_result != null) ...[
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surfaceContainerHighest,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Translation',
                      style: Theme.of(context).textTheme.labelSmall),
                  const SizedBox(height: 8),
                  SelectableText(_result!['translated_text']?.toString() ?? '',
                      style: Theme.of(context).textTheme.bodyLarge),
                  if (_result!['confidence'] != null) ...[
                    const SizedBox(height: 8),
                    Text(
                        'Confidence: ${((_result!['confidence'] ?? 0) * 100).toStringAsFixed(0)}%',
                        style: Theme.of(context).textTheme.bodySmall),
                  ],
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }
}

// ══════════════════════════════════════════════════
// TAB 5: DUPLICATES
// ══════════════════════════════════════════════════

class _DuplicateTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_DuplicateTab> createState() => _DuplicateTabState();
}

class _DuplicateTabState extends ConsumerState<_DuplicateTab> {
  final _ctrl = TextEditingController();
  Map<String, dynamic>? _result;
  bool _loading = false;

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  Future<void> _check() async {
    if (_ctrl.text.trim().isEmpty) return;
    setState(() => _loading = true);
    try {
      final api = ref.read(apiClientProvider);
      _result = await api.detectDuplicates(_ctrl.text.trim());
      setState(() {});
    } catch (e) {
      setState(() => _result = {'summary': 'Error: $e'});
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _FeatureHeader(
            icon: Icons.content_copy,
            color: const Color(0xFFFF9800),
            title: 'Duplicate Detection',
            subtitle:
                'Check if a report is similar to existing ones before submitting',
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _ctrl,
            maxLines: 4,
            decoration: InputDecoration(
              hintText: 'Enter report text to check for duplicates...',
              border:
                  OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            ),
          ),
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: _loading ? null : _check,
            icon: _loading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.search),
            label: Text(_loading ? 'Checking...' : 'Check Duplicates'),
            style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFFFF9800)),
          ),
          if (_result != null) ...[
            const SizedBox(height: 16),
            _ResultCard(
                title: 'Has Duplicates',
                value: _result!['has_duplicates'] == true ? '⚠️ Yes' : '✅ No'),
            _ResultCard(
                title: 'Recommendation',
                value: _result!['recommendation']?.toString() ?? 'N/A'),
            _ResultCard(
                title: 'Summary',
                value: _result!['summary']?.toString() ?? 'N/A'),
          ],
        ],
      ),
    );
  }
}

// ══════════════════════════════════════════════════
// TAB 7: ACTION PLAN
// ══════════════════════════════════════════════════

class _ActionPlanTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_ActionPlanTab> createState() => _ActionPlanTabState();
}

class _ActionPlanTabState extends ConsumerState<_ActionPlanTab> {
  final _ctrl = TextEditingController();
  Map<String, dynamic>? _result;
  bool _loading = false;

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  Future<void> _generate() async {
    final reportId = _ctrl.text.trim();
    if (reportId.isEmpty) return;
    setState(() => _loading = true);
    try {
      final api = ref.read(apiClientProvider);
      _result = await api.getActionPlan(reportId);
      setState(() {});
    } catch (e) {
      setState(() => _result = {'title': 'Error: $e'});
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _FeatureHeader(
            icon: Icons.task_alt,
            color: const Color(0xFF607D8B),
            title: 'AI Action Plan',
            subtitle: 'Generate field steps for an existing report ID',
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _ctrl,
            decoration: InputDecoration(
              hintText: 'Paste report ID',
              border:
                  OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
              prefixIcon: const Icon(Icons.confirmation_number),
            ),
            onSubmitted: (_) => _generate(),
          ),
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: _loading ? null : _generate,
            icon: _loading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.auto_awesome),
            label: Text(_loading ? 'Generating...' : 'Generate Action Plan'),
            style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFF607D8B)),
          ),
          if (_result != null) ...[
            const SizedBox(height: 16),
            _ResultCard(
                title: 'Title', value: _result!['title']?.toString() ?? 'N/A'),
            _ResultCard(
                title: 'Priority',
                value: _result!['priority_level']?.toString() ?? 'N/A'),
            _ResultCard(
                title: 'Estimated Duration',
                value: _result!['estimated_duration']?.toString() ?? 'N/A'),
            if (_result!['steps'] != null)
              _ResultCard(
                title: 'Steps',
                value: (_result!['steps'] as List)
                    .map((s) =>
                        '${s['step_number'] ?? '-'}: ${s['action'] ?? ''}\n${s['details'] ?? ''}')
                    .join('\n\n'),
              ),
            if (_result!['required_equipment'] != null)
              _ResultCard(
                  title: 'Equipment',
                  value: (_result!['required_equipment'] as List).join(', ')),
          ],
        ],
      ),
    );
  }
}

// ══════════════════════════════════════════════════
// TAB 8: OCR
// ══════════════════════════════════════════════════

class _OCRTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_OCRTab> createState() => _OCRTabState();
}

class _OCRTabState extends ConsumerState<_OCRTab> {
  Map<String, dynamic>? _result;
  bool _loading = false;
  Uint8List? _imageBytes;
  String? _imageName;
  String? _error;

  Future<void> _pickDocumentImage() async {
    final image = await ImagePicker().pickImage(
      source: ImageSource.gallery,
      maxWidth: 1600,
      imageQuality: 85,
    );
    if (image == null) return;
    final bytes = await image.readAsBytes();
    setState(() {
      _imageBytes = bytes;
      _imageName = image.name;
      _result = null;
      _error = null;
    });
  }

  Future<void> _scan() async {
    if (_imageBytes == null) {
      await _pickDocumentImage();
      if (_imageBytes == null) return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final api = ref.read(apiClientProvider);
      final payload = _imagePayload(_imageBytes!, _imageName);
      _result = await api.ocrDocument(payload);
      setState(() {});
    } catch (e) {
      setState(() => _error = 'OCR failed: $e');
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _FeatureHeader(
            icon: Icons.document_scanner,
            color: const Color(0xFF9C27B0),
            title: 'Document OCR',
            subtitle: 'Scan legal, medical, or government documents with AI',
          ),
          const SizedBox(height: 16),
          OutlinedButton.icon(
            onPressed: _loading ? null : _pickDocumentImage,
            icon: const Icon(Icons.upload_file),
            label: Text(_imageName == null ? 'Choose Document Image' : 'Change Document Image'),
          ),
          if (_imageBytes != null) ...[
            const SizedBox(height: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(12),
              child: Image.memory(_imageBytes!, height: 180, fit: BoxFit.cover),
            ),
            const SizedBox(height: 8),
            Text(_imageName ?? 'Selected document',
                textAlign: TextAlign.center,
                style: theme.textTheme.bodySmall),
          ],
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: _loading || _imageBytes == null ? null : _scan,
            icon: _loading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.document_scanner),
            label: Text(_loading ? 'Scanning...' : 'Scan Document'),
            style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFF9C27B0)),
          ),
          const SizedBox(height: 8),
          Text(
            'Capture or upload a document image to extract text',
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                ),
          ),
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(_error!, style: TextStyle(color: theme.colorScheme.error)),
          ],
          if (_result != null) ...[
            const SizedBox(height: 16),
            _ResultCard(
                title: 'Document Type',
                value: _result!['document_type']?.toString() ?? 'N/A'),
            _ResultCard(
                title: 'Language',
                value: _result!['language']?.toString() ?? 'N/A'),
            _ResultCard(
                title: 'Extracted Text',
                value: _result!['extracted_text']?.toString() ?? 'N/A'),
          ],
        ],
      ),
    );
  }
}

// ══════════════════════════════════════════════════
// TAB 7: PROGRESS REPORT
// ══════════════════════════════════════════════════

class _ProgressReportTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_ProgressReportTab> createState() => _ProgressReportTabState();
}

class _ProgressReportTabState extends ConsumerState<_ProgressReportTab> {
  Map<String, dynamic>? _report;
  bool _loading = false;

  Future<void> _generate() async {
    setState(() => _loading = true);
    try {
      final api = ref.read(apiClientProvider);
      _report = await api.getProgressReport();
      setState(() {});
    } catch (e) {
      setState(() => _report = {'executive_summary': 'Error: $e'});
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _FeatureHeader(
            icon: Icons.analytics,
            color: const Color(0xFF4CAF50),
            title: 'AI Progress Report',
            subtitle:
                'Generate a comprehensive weekly summary powered by Gemini',
          ),
          const SizedBox(height: 16),
          FilledButton.icon(
            onPressed: _loading ? null : _generate,
            icon: _loading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.auto_awesome),
            label: Text(_loading ? 'Generating...' : 'Generate Report'),
            style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFF4CAF50)),
          ),
          if (_report != null) ...[
            const SizedBox(height: 16),
            // Stats row
            Row(
              children: [
                _StatChip(
                    'Total', '${_report!['total_issues'] ?? 0}', Colors.blue),
                const SizedBox(width: 8),
                _StatChip('Resolved', '${_report!['resolved_issues'] ?? 0}',
                    Colors.green),
                const SizedBox(width: 8),
                _StatChip('Critical', '${_report!['critical_issues'] ?? 0}',
                    Colors.red),
              ],
            ),
            const SizedBox(height: 12),
            _ResultCard(
                title: 'Executive Summary',
                value: _report!['executive_summary']?.toString() ?? 'N/A'),
            if (_report!['key_achievements'] != null)
              _ResultCard(
                  title: 'Key Achievements',
                  value: (_report!['key_achievements'] as List).join('\n• ')),
            if (_report!['recommendations'] != null)
              _ResultCard(
                  title: 'Recommendations',
                  value: (_report!['recommendations'] as List).join('\n• ')),
          ],
        ],
      ),
    );
  }
}

// ══════════════════════════════════════════════════
// TAB 8: SKILL RECOMMENDATIONS
// ══════════════════════════════════════════════════

class _SkillsTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_SkillsTab> createState() => _SkillsTabState();
}

class _SkillsTabState extends ConsumerState<_SkillsTab> {
  Map<String, dynamic>? _result;
  bool _loading = false;

  Future<void> _getRecommendations() async {
    setState(() => _loading = true);
    try {
      final api = ref.read(apiClientProvider);
      _result = await api.getSkillRecommendations(
        ['general', 'first_aid'],
        ['civic_issue', 'medical_emergency'],
      );
      setState(() {});
    } catch (e) {
      setState(() => _result = {'career_path': 'Error: $e'});
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _FeatureHeader(
            icon: Icons.school,
            color: const Color(0xFFE91E63),
            title: 'AI Skill Recommendations',
            subtitle:
                'Get personalized skill development suggestions based on your task history',
          ),
          const SizedBox(height: 16),
          FilledButton.icon(
            onPressed: _loading ? null : _getRecommendations,
            icon: _loading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.lightbulb),
            label: Text(_loading ? 'Analyzing...' : 'Get Recommendations'),
            style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFFE91E63)),
          ),
          if (_result != null) ...[
            const SizedBox(height: 16),
            _ResultCard(
                title: 'Career Path',
                value: _result!['career_path']?.toString() ?? 'N/A'),
            if (_result!['recommended_skills'] != null)
              ...(_result!['recommended_skills'] as List).map(
                (s) => Card(
                  margin: const EdgeInsets.only(bottom: 8),
                  child: ListTile(
                    leading: Icon(Icons.star,
                        color: s['priority'] == 'high'
                            ? Colors.orange
                            : Colors.grey),
                    title: Text(s['skill']?.toString() ?? ''),
                    subtitle: Text(s['reason']?.toString() ?? ''),
                    trailing: Chip(
                        label: Text(s['priority']?.toString() ?? '',
                            style: const TextStyle(fontSize: 10))),
                  ),
                ),
              ),
            if (_result!['strengths'] != null)
              _ResultCard(
                  title: 'Your Strengths',
                  value: (_result!['strengths'] as List).join(', ')),
          ],
        ],
      ),
    );
  }
}

// ══════════════════════════════════════════════════
// SHARED COMPONENTS
// ══════════════════════════════════════════════════

class _FeatureHeader extends StatelessWidget {
  final IconData icon;
  final Color color;
  final String title;
  final String subtitle;

  const _FeatureHeader(
      {required this.icon,
      required this.color,
      required this.title,
      required this.subtitle});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(
            color: color.withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(icon, color: color, size: 28),
        ),
        const SizedBox(width: 14),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(title,
                  style: Theme.of(context)
                      .textTheme
                      .titleMedium
                      ?.copyWith(fontWeight: FontWeight.w700)),
              const SizedBox(height: 2),
              Text(subtitle,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      )),
            ],
          ),
        ),
      ],
    );
  }
}

class _ResultCard extends StatelessWidget {
  final String title;
  final String value;
  const _ResultCard({required this.title, required this.value});

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title,
                style: Theme.of(context).textTheme.labelSmall?.copyWith(
                      color: Theme.of(context).colorScheme.primary,
                      fontWeight: FontWeight.w600,
                    )),
            const SizedBox(height: 4),
            SelectableText(value,
                style: Theme.of(context).textTheme.bodyMedium),
          ],
        ),
      ),
    );
  }
}

class _StatChip extends StatelessWidget {
  final String label;
  final String value;
  final Color color;
  const _StatChip(this.label, this.value, this.color);

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: color.withValues(alpha: 0.3)),
        ),
        child: Column(
          children: [
            Text(value,
                style: TextStyle(
                    fontSize: 24, fontWeight: FontWeight.w800, color: color)),
            Text(label, style: TextStyle(fontSize: 11, color: color)),
          ],
        ),
      ),
    );
  }
}
