import '../../data/local/hive_store.dart';
import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import '../../main.dart';

/// Specialist (Lawyer/Doctor) Dashboard — Case files, document upload, AI semantic search.
class SpecialistDashboard extends ConsumerStatefulWidget {
  const SpecialistDashboard({super.key});

  @override
  ConsumerState<SpecialistDashboard> createState() => _SpecialistDashboardState();
}

class _SpecialistDashboardState extends ConsumerState<SpecialistDashboard> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  List<Map<String, dynamic>> _cases = [];
  bool _isLoading = true;

  // AI Chat state
  final _chatController = TextEditingController();
  final List<_ChatMessage> _chatMessages = [];
  bool _isChatLoading = false;
  String? _selectedCaseId;

  // Document search state
  final _docSearchController = TextEditingController();
  List<Map<String, dynamic>> _docSearchResults = [];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _loadCases();
  }

  @override
  void dispose() {
    _tabController.dispose();
    _chatController.dispose();
    _docSearchController.dispose();
    super.dispose();
  }

  Future<void> _loadCases() async {
    setState(() => _isLoading = true);
    try {
      final api = ref.read(apiClientProvider);
      final cases = await api.getMyCases();
      setState(() {
        _cases = List<Map<String, dynamic>>.from(cases);
        if (_cases.isNotEmpty && _selectedCaseId == null) {
          _selectedCaseId = _cases.first['id'] as String?;
        }
        _isLoading = false;
      });
    } catch (e) {
      debugPrint('Load cases error: $e');
      setState(() => _isLoading = false);
    }
  }

  Future<void> _askQuestion() async {
    if (_chatController.text.isEmpty || _selectedCaseId == null) return;

    final question = _chatController.text;
    _chatController.clear();

    setState(() {
      _chatMessages.add(_ChatMessage(text: question, isUser: true));
      _isChatLoading = true;
    });

    try {
      final api = ref.read(apiClientProvider);
      final response = await api.askCaseQuestion(_selectedCaseId!, question);
      setState(() {
        _chatMessages.add(_ChatMessage(text: response['answer'] as String? ?? 'No answer found.', isUser: false));
        _isChatLoading = false;
      });
    } catch (e) {
      setState(() {
        _chatMessages.add(_ChatMessage(text: 'Error: Could not get answer. Make sure documents are uploaded.', isUser: false));
        _isChatLoading = false;
      });
    }
  }

  Future<void> _uploadDocument() async {
    if (_selectedCaseId == null) {
      _showSnackBar('Select a case first');
      return;
    }

    final picker = ImagePicker();
    final file = await picker.pickImage(source: ImageSource.gallery, maxWidth: 1200, imageQuality: 60);
    if (file == null) return;

    final bytes = await file.readAsBytes();
    final base64Content = 'data:image/${file.name.split('.').last};base64,${base64Encode(bytes)}';

    try {
      final api = ref.read(apiClientProvider);
      await api.uploadCaseDocument(_selectedCaseId!, file.name, base64Content, 'image');
      _showSnackBar('Document uploaded & indexed');
      _loadCases();
    } catch (e) {
      _showSnackBar('Upload failed');
    }
  }

  void _showSnackBar(String msg) {
    if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Specialist Mode'),
        actions: [
          IconButton(
            icon: const Icon(Icons.auto_awesome),
            tooltip: 'AI Hub',
            onPressed: () => context.push('/ai-hub'),
          ),
          IconButton(icon: const Icon(Icons.refresh), onPressed: _loadCases),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await HiveStore.clearAll();
              await FirebaseAuth.instance.signOut();
              if (mounted) context.go('/role-select');
            },
          ),
        ],
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(icon: Icon(Icons.folder_special), text: 'My Cases'),
            Tab(icon: Icon(Icons.psychology), text: 'AI Chat'),
            Tab(icon: Icon(Icons.search), text: 'Doc Search'),
          ],
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : TabBarView(
              controller: _tabController,
              children: [
                _buildCasesTab(theme),
                _buildAIChatTab(theme),
                _buildDocSearchTab(theme),
              ],
            ),
    );
  }

  // --- Tab 1: My Cases ---
  Widget _buildCasesTab(ThemeData theme) {
    if (_cases.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.folder_off, size: 64, color: theme.colorScheme.outlineVariant),
            const SizedBox(height: 16),
            Text('No cases assigned yet', style: theme.textTheme.bodyLarge),
            Text('Cases will appear when an NGO admin assigns you.',
                style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: _cases.length,
      itemBuilder: (context, index) {
        final caseFile = _cases[index];
        final docs = (caseFile['documents'] as List?)?.length ?? 0;
        final status = (caseFile['status'] as String?) ?? 'open';

        return Card(
          margin: const EdgeInsets.only(bottom: 12),
          child: InkWell(
            onTap: () {
              setState(() => _selectedCaseId = caseFile['id'] as String?);
              _tabController.animateTo(1); // Switch to AI Chat
            },
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.folder_special, color: theme.colorScheme.primary),
                      const SizedBox(width: 8),
                      Expanded(child: Text(
                        (caseFile['title'] as String?) ?? 'Case #${caseFile['id']}',
                        style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w600),
                      )),
                      Chip(
                        label: Text(status.toUpperCase(), style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold,
                          color: status == 'open' ? Colors.blue : Colors.green)),
                        backgroundColor: (status == 'open' ? Colors.blue : Colors.green).withOpacity(0.1),
                        side: BorderSide.none, padding: EdgeInsets.zero,
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text('$docs documents attached',
                      style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      OutlinedButton.icon(
                        onPressed: _uploadDocument,
                        icon: const Icon(Icons.upload_file, size: 16),
                        label: const Text('Add Document'),
                      ),
                      const SizedBox(width: 8),
                      FilledButton.icon(
                        onPressed: () {
                          setState(() => _selectedCaseId = caseFile['id'] as String?);
                          _tabController.animateTo(1);
                        },
                        icon: const Icon(Icons.psychology, size: 16),
                        label: const Text('Ask AI'),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  // --- Tab 2: AI Chat ---
  Widget _buildAIChatTab(ThemeData theme) {
    return Column(
      children: [
        // Case selector
        if (_cases.isNotEmpty)
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            color: theme.colorScheme.surfaceContainerLow,
            child: Row(
              children: [
                Text('Case: ', style: theme.textTheme.labelMedium),
                Expanded(
                  child: DropdownButton<String>(
                    value: _selectedCaseId,
                    isExpanded: true,
                    underline: const SizedBox(),
                    items: _cases.map((c) => DropdownMenuItem(
                      value: c['id'] as String?,
                      child: Text((c['title'] as String?) ?? 'Case ${c['id']}', overflow: TextOverflow.ellipsis),
                    )).toList(),
                    onChanged: (v) => setState(() => _selectedCaseId = v),
                  ),
                ),
              ],
            ),
          ),

        // Chat messages
        Expanded(
          child: _chatMessages.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.psychology, size: 64, color: theme.colorScheme.outlineVariant),
                      const SizedBox(height: 16),
                      Text('Ask questions about your case documents', style: theme.textTheme.bodyLarge),
                      const SizedBox(height: 4),
                      Text('Powered by Gemini AI semantic search',
                          style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: _chatMessages.length + (_isChatLoading ? 1 : 0),
                  itemBuilder: (context, index) {
                    if (index == _chatMessages.length) {
                      return const Padding(
                        padding: EdgeInsets.all(16),
                        child: Center(child: CircularProgressIndicator()),
                      );
                    }
                    final msg = _chatMessages[index];
                    return Align(
                      alignment: msg.isUser ? Alignment.centerRight : Alignment.centerLeft,
                      child: Container(
                        margin: const EdgeInsets.only(bottom: 8),
                        padding: const EdgeInsets.all(14),
                        constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.75),
                        decoration: BoxDecoration(
                          color: msg.isUser ? theme.colorScheme.primary : theme.colorScheme.surfaceContainerHighest,
                          borderRadius: BorderRadius.circular(16),
                        ),
                        child: Text(msg.text, style: TextStyle(
                          color: msg.isUser ? theme.colorScheme.onPrimary : theme.colorScheme.onSurface)),
                      ),
                    );
                  },
                ),
        ),

        // Chat input
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: theme.colorScheme.surface,
            boxShadow: [BoxShadow(color: Colors.black12, blurRadius: 4, offset: const Offset(0, -2))],
          ),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _chatController,
                  onSubmitted: (_) => _askQuestion(),
                  decoration: InputDecoration(
                    hintText: 'Ask about your case documents...',
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(24)),
                    contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              IconButton.filled(
                onPressed: _askQuestion,
                icon: const Icon(Icons.send),
              ),
            ],
          ),
        ),
      ],
    );
  }

  // --- Tab 3: Document Search ---
  Widget _buildDocSearchTab(ThemeData theme) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(16),
          child: TextField(
            controller: _docSearchController,
            onSubmitted: (_) => _performDocSearch(),
            decoration: InputDecoration(
              hintText: 'Search within a specific document...',
              prefixIcon: const Icon(Icons.search),
              suffixIcon: IconButton(icon: const Icon(Icons.send), onPressed: _performDocSearch),
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            ),
          ),
        ),
        Expanded(
          child: _docSearchResults.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.find_in_page, size: 64, color: theme.colorScheme.outlineVariant),
                      const SizedBox(height: 16),
                      Text('Search inside documents', style: theme.textTheme.bodyLarge),
                      Text('Find specific sections within a single document',
                          style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  itemCount: _docSearchResults.length,
                  itemBuilder: (context, index) {
                    final result = _docSearchResults[index];
                    return Card(
                      margin: const EdgeInsets.only(bottom: 8),
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(result['file_name'] as String? ?? 'Document',
                                style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w600)),
                            const SizedBox(height: 4),
                            Text(result['excerpt'] as String? ?? '',
                                style: theme.textTheme.bodySmall),
                            const SizedBox(height: 4),
                            Text('Relevance: ${((result['score'] as double?) ?? 0 * 100).toStringAsFixed(0)}%',
                                style: theme.textTheme.labelSmall?.copyWith(color: theme.colorScheme.primary)),
                          ],
                        ),
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }

  Future<void> _performDocSearch() async {
    if (_docSearchController.text.isEmpty || _selectedCaseId == null) return;
    try {
      final api = ref.read(apiClientProvider);
      final results = await api.searchCaseDocuments(_selectedCaseId!, _docSearchController.text);
      setState(() => _docSearchResults = List<Map<String, dynamic>>.from(results));
    } catch (e) {
      _showSnackBar('Search failed');
    }
  }
}

class _ChatMessage {
  final String text;
  final bool isUser;
  const _ChatMessage({required this.text, required this.isUser});
}
