import 'package:flutter/material.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:go_router/go_router.dart';

/// Lawyer "Legal Mode" — Gemini-powered semantic search across FIRs and land deeds.
class LegalDashboard extends StatefulWidget {
  const LegalDashboard({super.key});

  @override
  State<LegalDashboard> createState() => _LegalDashboardState();
}

class _LegalDashboardState extends State<LegalDashboard> {
  final _searchController = TextEditingController();
  String _selectedCategory = 'All';
  List<Map<String, dynamic>> _results = [];
  bool _isSearching = false;

  static const _categories = ['All', 'FIR', 'Land Deed', 'Court Order', 'Affidavit'];

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _performSearch() async {
    if (_searchController.text.isEmpty) return;

    setState(() => _isSearching = true);

    // TODO: Wire to API — Gemini semantic search endpoint
    await Future.delayed(const Duration(seconds: 1));

    setState(() {
      _isSearching = false;
      _results = [
        {
          'title': 'FIR #2024/1847 — Water Contamination',
          'type': 'FIR',
          'date': '2024-12-15',
          'relevance': 0.94,
          'summary': 'Filed at Gomti Nagar PS regarding contaminated water supply in Ward 23.',
        },
        {
          'title': 'Land Deed #LK-4521 — Disputed Plot',
          'type': 'Land Deed',
          'date': '2023-08-20',
          'relevance': 0.87,
          'summary': 'Transfer deed for plot 45/A in Chinhat tehsil, disputed ownership.',
        },
      ];
    });
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Legal Mode'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await FirebaseAuth.instance.signOut();
              if (mounted) context.go('/login');
            },
          ),
        ],
      ),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              children: [
                // Search bar
                TextField(
                  controller: _searchController,
                  onSubmitted: (_) => _performSearch(),
                  decoration: InputDecoration(
                    hintText: 'Search FIRs, deeds, legal documents...',
                    prefixIcon: const Icon(Icons.search),
                    suffixIcon: IconButton(
                      icon: const Icon(Icons.send_rounded),
                      onPressed: _performSearch,
                    ),
                  ),
                ),
                const SizedBox(height: 12),
                // Category chips
                SizedBox(
                  height: 36,
                  child: ListView.separated(
                    scrollDirection: Axis.horizontal,
                    itemCount: _categories.length,
                    separatorBuilder: (_, __) => const SizedBox(width: 8),
                    itemBuilder: (context, index) {
                      final cat = _categories[index];
                      final isSelected = cat == _selectedCategory;
                      return FilterChip(
                        label: Text(cat),
                        selected: isSelected,
                        onSelected: (v) => setState(() => _selectedCategory = cat),
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
          if (_isSearching)
            const Expanded(child: Center(child: CircularProgressIndicator()))
          else
            Expanded(
              child: _results.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.gavel_rounded, size: 64,
                              color: theme.colorScheme.outlineVariant),
                          const SizedBox(height: 16),
                          Text('Search across digitized legal records',
                              style: theme.textTheme.bodyLarge),
                          Text('Powered by Gemini AI semantic search',
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant)),
                        ],
                      ),
                    )
                  : ListView.builder(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      itemCount: _results.length,
                      itemBuilder: (context, index) {
                        final doc = _results[index];
                        return Card(
                          margin: const EdgeInsets.only(bottom: 12),
                          child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  children: [
                                    Chip(label: Text(doc['type'] as String)),
                                    const Spacer(),
                                    Text(
                                      '${((doc['relevance'] as double) * 100).round()}% match',
                                      style: theme.textTheme.labelMedium?.copyWith(
                                        color: theme.colorScheme.primary,
                                        fontWeight: FontWeight.w600,
                                      ),
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 8),
                                Text(doc['title'] as String,
                                    style: theme.textTheme.titleSmall?.copyWith(
                                        fontWeight: FontWeight.w600)),
                                const SizedBox(height: 4),
                                Text(doc['summary'] as String,
                                    style: theme.textTheme.bodySmall),
                                const SizedBox(height: 8),
                                Text('Filed: ${doc['date']}',
                                    style: theme.textTheme.labelSmall?.copyWith(
                                      color: theme.colorScheme.onSurfaceVariant)),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
            ),
        ],
      ),
    );
  }
}
