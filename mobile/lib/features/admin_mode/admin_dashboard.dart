import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:go_router/go_router.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import '../../main.dart';
import '../../data/local/hive_store.dart';

/// NGO Admin Dashboard — Issue management, volunteer matching, heatmap.
class AdminDashboard extends ConsumerStatefulWidget {
  const AdminDashboard({super.key});

  @override
  ConsumerState<AdminDashboard> createState() => _AdminDashboardState();
}

class _AdminDashboardState extends ConsumerState<AdminDashboard>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  List<Map<String, dynamic>> _reports = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _loadData();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);
    try {
      final api = ref.read(apiClientProvider);
      final data = await api.getPrioritizedReports();
      setState(() {
        _reports = List<Map<String, dynamic>>.from(data);
        _isLoading = false;
      });
    } catch (e) {
      debugPrint('Load reports error: $e');
      setState(() => _isLoading = false);
    }
  }

  Future<void> _assignVolunteers(String reportId) async {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (ctx) => _VolunteerMatchSheet(
        reportId: reportId,
        api: ref.read(apiClientProvider),
        onAssigned: () {
          Navigator.pop(ctx);
          _loadData();
          _showSnackBar('Volunteers assigned successfully!');
        },
      ),
    );
  }

  void _showSnackBar(String msg) {
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(
        content: Text(msg),
        behavior: SnackBarBehavior.floating,
      ));
    }
  }

  Color _urgencyColor(String urgency) {
    switch (urgency) {
      case 'critical':
        return const Color(0xFFFF1744);
      case 'urgent':
        return const Color(0xFFFF9100);
      default:
        return const Color(0xFF00E676);
    }
  }

  Color _issueTypeColor(String type) {
    switch (type) {
      case 'medical_emergency':
        return const Color(0xFFFF1744);
      case 'legal_aid':
        return const Color(0xFF448AFF);
      case 'disaster_relief':
        return const Color(0xFFFF6D00);
      default:
        return const Color(0xFF00BFA5);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Scaffold(
      appBar: AppBar(
        title: const Text('NGO Command Center'),
        actions: [
          IconButton(
            icon: const Icon(Icons.auto_awesome),
            tooltip: 'AI Hub',
            onPressed: () => context.push('/ai-hub'),
          ),
          IconButton(icon: const Icon(Icons.refresh), onPressed: _loadData),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await HiveStore.clearAll();
              await FirebaseAuth.instance.signOut();
              if (!context.mounted) return;
              context.go('/role-select');
            },
          ),
        ],
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(icon: Icon(Icons.list_alt), text: 'Issues'),
            Tab(icon: Icon(Icons.people_alt), text: 'Matching'),
            Tab(icon: Icon(Icons.map_outlined), text: 'Heatmap'),
          ],
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : TabBarView(
              controller: _tabController,
              children: [
                _buildIssuesTab(theme, isDark),
                _buildMatchingTab(theme),
                _HeatmapTab(reports: _reports),
              ],
            ),
    );
  }

  // ── Tab 1: Prioritized Issues ──
  Widget _buildIssuesTab(ThemeData theme, bool isDark) {
    if (_reports.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.check_circle_outline,
                size: 64, color: theme.colorScheme.outlineVariant),
            const SizedBox(height: 16),
            Text('No issues reported yet', style: theme.textTheme.bodyLarge),
          ],
        ),
      );
    }

    final critical =
        _reports.where((r) => r['user_urgency'] == 'critical').length;
    final urgent = _reports.where((r) => r['user_urgency'] == 'urgent').length;
    final pending = _reports.where((r) => r['status'] == 'pending').length;

    return Column(
      children: [
        // Stats bar
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: isDark ? const Color(0xFF1A1A2E) : Colors.white,
            border: Border(bottom: BorderSide(color: theme.dividerColor)),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _StatChip(
                  label: 'Total',
                  value: '${_reports.length}',
                  color: const Color(0xFF448AFF)),
              _StatChip(
                  label: 'Critical',
                  value: '$critical',
                  color: const Color(0xFFFF1744)),
              _StatChip(
                  label: 'Urgent',
                  value: '$urgent',
                  color: const Color(0xFFFF9100)),
              _StatChip(
                  label: 'Pending', value: '$pending', color: Colors.grey),
            ],
          ),
        ),
        Expanded(
          child: RefreshIndicator(
            onRefresh: _loadData,
            child: ListView.builder(
              padding: const EdgeInsets.all(12),
              itemCount: _reports.length,
              itemBuilder: (context, index) {
                final report = _reports[index];
                final urgency = (report['user_urgency'] as String?) ?? 'normal';
                final issueType =
                    (report['issue_type'] as String?) ?? 'civic_issue';
                final status = (report['status'] as String?) ?? 'pending';
                final reqVol =
                    (report['required_volunteers'] as num?)?.toInt() ?? 1;
                final assigned =
                    (report['assigned_volunteer_ids'] as List?)?.length ?? 0;

                return Card(
                  margin: const EdgeInsets.only(bottom: 10),
                  child: Padding(
                    padding: const EdgeInsets.all(14),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Priority header
                        Row(
                          children: [
                            Container(
                              width: 30,
                              height: 30,
                              decoration: BoxDecoration(
                                color: _urgencyColor(urgency)
                                    .withValues(alpha: 0.15),
                                shape: BoxShape.circle,
                              ),
                              child: Center(
                                  child: Text('#${index + 1}',
                                      style: TextStyle(
                                          fontWeight: FontWeight.bold,
                                          fontSize: 11,
                                          color: _urgencyColor(urgency)))),
                            ),
                            const SizedBox(width: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 8, vertical: 3),
                              decoration: BoxDecoration(
                                color: _issueTypeColor(issueType)
                                    .withValues(alpha: 0.12),
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Text(issueType.replaceAll('_', ' '),
                                  style: TextStyle(
                                      fontSize: 10,
                                      fontWeight: FontWeight.w600,
                                      color: _issueTypeColor(issueType))),
                            ),
                            const SizedBox(width: 6),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 8, vertical: 3),
                              decoration: BoxDecoration(
                                color: _urgencyColor(urgency)
                                    .withValues(alpha: 0.12),
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Text(urgency.toUpperCase(),
                                  style: TextStyle(
                                      fontSize: 10,
                                      fontWeight: FontWeight.bold,
                                      color: _urgencyColor(urgency))),
                            ),
                            const Spacer(),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 8, vertical: 3),
                              decoration: BoxDecoration(
                                color: status == 'resolved'
                                    ? Colors.green.withValues(alpha: 0.1)
                                    : Colors.grey.withValues(alpha: 0.1),
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Text(status.replaceAll('_', ' '),
                                  style: theme.textTheme.labelSmall?.copyWith(
                                    fontWeight: FontWeight.w600,
                                    color: status == 'resolved'
                                        ? Colors.green
                                        : theme.colorScheme.onSurfaceVariant,
                                  )),
                            ),
                          ],
                        ),
                        const SizedBox(height: 10),
                        Text(
                            (report['raw_text'] as String?) ?? 'No description',
                            style: theme.textTheme.bodyMedium,
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis),
                        const SizedBox(height: 8),
                        Row(
                          children: [
                            if ((report['location'] as String?)?.isNotEmpty ==
                                true) ...[
                              Icon(Icons.location_on,
                                  size: 14,
                                  color: theme.colorScheme.onSurfaceVariant),
                              const SizedBox(width: 2),
                              Expanded(
                                  child: Text(
                                      report['location'] as String? ?? '',
                                      style: theme.textTheme.labelSmall,
                                      overflow: TextOverflow.ellipsis)),
                            ],
                            Icon(Icons.people,
                                size: 14,
                                color: theme.colorScheme.onSurfaceVariant),
                            const SizedBox(width: 2),
                            Text('$assigned/$reqVol assigned',
                                style: theme.textTheme.labelSmall),
                          ],
                        ),
                        const SizedBox(height: 10),
                        SizedBox(
                          width: double.infinity,
                          child: FilledButton.icon(
                            onPressed: () =>
                                _assignVolunteers(report['id'] as String),
                            icon: const Icon(Icons.person_add_alt_1, size: 16),
                            label: const Text('Find & Assign'),
                          ),
                        ),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
        ),
      ],
    );
  }

  // ── Tab 2: Volunteer Matching ──
  Widget _buildMatchingTab(ThemeData theme) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.people_alt_rounded,
                size: 64, color: Color(0xFF6C63FF)),
            const SizedBox(height: 16),
            Text('Volunteer Matching',
                style: theme.textTheme.titleLarge
                    ?.copyWith(fontWeight: FontWeight.w700)),
            const SizedBox(height: 8),
            Text(
                'Select an issue from the Issues tab and tap "Find & Assign" to match volunteers.',
                textAlign: TextAlign.center,
                style: theme.textTheme.bodyMedium
                    ?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
            const SizedBox(height: 24),
            FilledButton.icon(
              onPressed: () => _tabController.animateTo(0),
              icon: const Icon(Icons.list_alt),
              label: const Text('Go to Issues'),
            ),
          ],
        ),
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════
// GOOGLE MAPS HEATMAP TAB
// ═══════════════════════════════════════════════════════

class _HeatmapTab extends StatefulWidget {
  final List<Map<String, dynamic>> reports;
  const _HeatmapTab({required this.reports});

  @override
  State<_HeatmapTab> createState() => _HeatmapTabState();
}

class _HeatmapTabState extends State<_HeatmapTab> {
  GoogleMapController? _mapController;
  MapType _mapType = MapType.normal;

  @override
  void didUpdateWidget(covariant _HeatmapTab oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.reports != widget.reports) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _fitAllReports());
    }
  }

  List<LatLng> _validReportPoints() {
    final points = <LatLng>[];
    for (final r in widget.reports) {
      final lat = (r['latitude'] as num?)?.toDouble();
      final lng = (r['longitude'] as num?)?.toDouble();
      if (lat == null || lng == null || (lat == 0 && lng == 0)) continue;
      if (lat < -90 || lat > 90 || lng < -180 || lng > 180) continue;
      points.add(LatLng(lat, lng));
    }
    return points;
  }

  Future<void> _fitAllReports() async {
    final controller = _mapController;
    if (controller == null) return;

    final points = _validReportPoints();
    if (points.isEmpty) return;
    if (points.length == 1) {
      await controller
          .animateCamera(CameraUpdate.newLatLngZoom(points.first, 15));
      return;
    }

    var minLat = points.first.latitude;
    var maxLat = points.first.latitude;
    var minLng = points.first.longitude;
    var maxLng = points.first.longitude;
    for (final point in points.skip(1)) {
      if (point.latitude < minLat) minLat = point.latitude;
      if (point.latitude > maxLat) maxLat = point.latitude;
      if (point.longitude < minLng) minLng = point.longitude;
      if (point.longitude > maxLng) maxLng = point.longitude;
    }

    await controller.animateCamera(CameraUpdate.newLatLngBounds(
      LatLngBounds(
        southwest: LatLng(minLat, minLng),
        northeast: LatLng(maxLat, maxLng),
      ),
      80,
    ));
  }

  LatLng _computeCenter() {
    final points = _validReportPoints();
    if (points.isEmpty) {
      return const LatLng(26.8467, 80.9462); // Default: Lucknow
    }
    final sumLat = points.fold<double>(0, (sum, point) => sum + point.latitude);
    final sumLng =
        points.fold<double>(0, (sum, point) => sum + point.longitude);
    return LatLng(sumLat / points.length, sumLng / points.length);
  }

  Set<Marker> _buildMarkers() {
    final markers = <Marker>{};
    for (int i = 0; i < widget.reports.length; i++) {
      final r = widget.reports[i];
      final lat = (r['latitude'] as num?)?.toDouble();
      final lng = (r['longitude'] as num?)?.toDouble();
      if (lat == null || lng == null || (lat == 0 && lng == 0)) continue;

      final urgency = (r['user_urgency'] as String?) ?? 'normal';
      final issueType = (r['issue_type'] as String?) ?? 'civic_issue';
      final rawText = (r['raw_text'] as String?) ?? 'No description';
      final location = (r['location'] as String?) ?? '';
      final snippet =
          rawText.length > 60 ? '${rawText.substring(0, 60)}...' : rawText;

      // Choose marker color based on urgency
      BitmapDescriptor markerIcon;
      switch (urgency) {
        case 'critical':
          markerIcon =
              BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueRed);
          break;
        case 'urgent':
          markerIcon =
              BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueOrange);
          break;
        default:
          // Color by issue type
          switch (issueType) {
            case 'medical_emergency':
              markerIcon = BitmapDescriptor.defaultMarkerWithHue(
                  BitmapDescriptor.hueRose);
              break;
            case 'legal_aid':
              markerIcon = BitmapDescriptor.defaultMarkerWithHue(
                  BitmapDescriptor.hueBlue);
              break;
            case 'disaster_relief':
              markerIcon = BitmapDescriptor.defaultMarkerWithHue(
                  BitmapDescriptor.hueYellow);
              break;
            default:
              markerIcon = BitmapDescriptor.defaultMarkerWithHue(
                  BitmapDescriptor.hueGreen);
          }
      }

      markers.add(Marker(
        markerId: MarkerId((r['id'] as String?) ?? 'report_$i'),
        position: LatLng(lat, lng),
        icon: markerIcon,
        infoWindow: InfoWindow(
          title: '${urgency.toUpperCase()} — ${issueType.replaceAll('_', ' ')}',
          snippet: location.isNotEmpty ? '$snippet\n📍 $location' : snippet,
        ),
      ));
    }
    return markers;
  }

  Set<Circle> _buildHeatCircles() {
    final circles = <Circle>{};
    for (int i = 0; i < widget.reports.length; i++) {
      final r = widget.reports[i];
      final lat = (r['latitude'] as num?)?.toDouble();
      final lng = (r['longitude'] as num?)?.toDouble();
      if (lat == null || lng == null || (lat == 0 && lng == 0)) continue;

      final urgency = (r['user_urgency'] as String?) ?? 'normal';
      final radius =
          urgency == 'critical' ? 500.0 : (urgency == 'urgent' ? 350.0 : 200.0);

      Color fillColor;
      switch (urgency) {
        case 'critical':
          fillColor = Colors.red.withValues(alpha: 0.2);
          break;
        case 'urgent':
          fillColor = Colors.orange.withValues(alpha: 0.18);
          break;
        default:
          fillColor = Colors.green.withValues(alpha: 0.15);
      }

      circles.add(Circle(
        circleId: CircleId('heat_$i'),
        center: LatLng(lat, lng),
        radius: radius,
        fillColor: fillColor,
        strokeColor: fillColor.withValues(alpha: 0.5),
        strokeWidth: 1,
      ));
    }
    return circles;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;
    final center = _computeCenter();
    final markers = _buildMarkers();
    final circles = _buildHeatCircles();

    final totalReports = widget.reports.length;
    final criticalCount =
        widget.reports.where((r) => r['user_urgency'] == 'critical').length;
    final urgentCount =
        widget.reports.where((r) => r['user_urgency'] == 'urgent').length;
    final pendingCount =
        widget.reports.where((r) => r['status'] == 'pending').length;

    return Stack(
      children: [
        // Google Map
        GoogleMap(
          initialCameraPosition: CameraPosition(target: center, zoom: 12),
          onMapCreated: (controller) {
            _mapController = controller;
            WidgetsBinding.instance
                .addPostFrameCallback((_) => _fitAllReports());
          },
          style: isDark ? _darkMapStyle : null,
          markers: markers,
          circles: circles,
          mapType: _mapType,
          myLocationEnabled: false,
          zoomControlsEnabled: false,
          mapToolbarEnabled: false,
        ),

        // Top legend
        Positioned(
          top: 12,
          left: 12,
          right: 12,
          child: Card(
            elevation: 4,
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                children: [
                  _LegendItem(
                      color: const Color(0xFFFF1744),
                      label: 'Critical ($criticalCount)'),
                  _LegendItem(
                      color: const Color(0xFFFF9100),
                      label: 'Urgent ($urgentCount)'),
                  _LegendItem(color: const Color(0xFF00E676), label: 'Normal'),
                  Text('$totalReports total',
                      style: theme.textTheme.labelSmall
                          ?.copyWith(fontWeight: FontWeight.bold)),
                ],
              ),
            ),
          ),
        ),

        // Map type toggle
        Positioned(
          top: 70,
          right: 12,
          child: Column(
            children: [
              _MapButton(
                icon: Icons.layers,
                onTap: () => setState(() {
                  _mapType = _mapType == MapType.normal
                      ? MapType.satellite
                      : MapType.normal;
                }),
              ),
              const SizedBox(height: 8),
              _MapButton(
                icon: Icons.my_location,
                onTap: () => _mapController
                    ?.animateCamera(CameraUpdate.newLatLngZoom(center, 12)),
              ),
              const SizedBox(height: 8),
              _MapButton(
                icon: Icons.zoom_in,
                onTap: () =>
                    _mapController?.animateCamera(CameraUpdate.zoomIn()),
              ),
              const SizedBox(height: 8),
              _MapButton(
                icon: Icons.zoom_out,
                onTap: () =>
                    _mapController?.animateCamera(CameraUpdate.zoomOut()),
              ),
            ],
          ),
        ),

        // Bottom stats
        Positioned(
          bottom: 0,
          left: 0,
          right: 0,
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
            decoration: BoxDecoration(
              color: isDark ? const Color(0xFF161628) : Colors.white,
              boxShadow: [
                BoxShadow(
                    color: Colors.black.withValues(alpha: 0.1),
                    blurRadius: 10,
                    offset: const Offset(0, -2))
              ],
              borderRadius:
                  const BorderRadius.vertical(top: Radius.circular(16)),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _BottomStat(
                    icon: Icons.warning_amber_rounded,
                    value: '$totalReports',
                    label: 'Issues',
                    color: const Color(0xFF448AFF)),
                _BottomStat(
                    icon: Icons.error_outline,
                    value: '$criticalCount',
                    label: 'Critical',
                    color: const Color(0xFFFF1744)),
                _BottomStat(
                    icon: Icons.schedule,
                    value: '$pendingCount',
                    label: 'Pending',
                    color: const Color(0xFFFF9100)),
                _BottomStat(
                    icon: Icons.check_circle_outline,
                    value: '${totalReports - pendingCount}',
                    label: 'Active',
                    color: const Color(0xFF00E676)),
              ],
            ),
          ),
        ),
      ],
    );
  }
}

// Dark mode map style
const String _darkMapStyle = '''[
  {"elementType": "geometry", "stylers": [{"color": "#212121"}]},
  {"elementType": "labels.icon", "stylers": [{"visibility": "off"}]},
  {"elementType": "labels.text.fill", "stylers": [{"color": "#757575"}]},
  {"elementType": "labels.text.stroke", "stylers": [{"color": "#212121"}]},
  {"featureType": "administrative", "elementType": "geometry", "stylers": [{"color": "#757575"}]},
  {"featureType": "poi", "elementType": "geometry", "stylers": [{"color": "#181818"}]},
  {"featureType": "poi.park", "elementType": "geometry", "stylers": [{"color": "#1b3a1b"}]},
  {"featureType": "road", "elementType": "geometry.fill", "stylers": [{"color": "#2c2c2c"}]},
  {"featureType": "road", "elementType": "labels.text.fill", "stylers": [{"color": "#8a8a8a"}]},
  {"featureType": "road.highway", "elementType": "geometry", "stylers": [{"color": "#3c3c3c"}]},
  {"featureType": "transit", "elementType": "geometry", "stylers": [{"color": "#2f3948"}]},
  {"featureType": "water", "elementType": "geometry", "stylers": [{"color": "#000000"}]},
  {"featureType": "water", "elementType": "labels.text.fill", "stylers": [{"color": "#3d3d3d"}]}
]''';

// ═══════════════════════════════════════════════════════
// HELPER WIDGETS
// ═══════════════════════════════════════════════════════

class _StatChip extends StatelessWidget {
  final String label;
  final String value;
  final Color color;
  const _StatChip(
      {required this.label, required this.value, required this.color});

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(value,
            style: Theme.of(context)
                .textTheme
                .titleMedium
                ?.copyWith(fontWeight: FontWeight.bold, color: color)),
        const SizedBox(height: 2),
        Text(label, style: Theme.of(context).textTheme.labelSmall),
      ],
    );
  }
}

class _LegendItem extends StatelessWidget {
  final Color color;
  final String label;
  const _LegendItem({required this.color, required this.label});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
            width: 10,
            height: 10,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle)),
        const SizedBox(width: 4),
        Text(label, style: Theme.of(context).textTheme.labelSmall),
      ],
    );
  }
}

class _MapButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback onTap;
  const _MapButton({required this.icon, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Material(
      elevation: 3,
      borderRadius: BorderRadius.circular(8),
      color: isDark ? const Color(0xFF1E1E32) : Colors.white,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Icon(icon, size: 20),
        ),
      ),
    );
  }
}

class _BottomStat extends StatelessWidget {
  final IconData icon;
  final String value;
  final String label;
  final Color color;
  const _BottomStat(
      {required this.icon,
      required this.value,
      required this.label,
      required this.color});

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 20, color: color),
        const SizedBox(height: 4),
        Text(value,
            style: Theme.of(context)
                .textTheme
                .titleSmall
                ?.copyWith(fontWeight: FontWeight.bold)),
        Text(label, style: Theme.of(context).textTheme.labelSmall),
      ],
    );
  }
}

class _VolunteerMatchSheet extends StatefulWidget {
  final String reportId;
  final dynamic api;
  final VoidCallback onAssigned;

  const _VolunteerMatchSheet(
      {required this.reportId, required this.api, required this.onAssigned});

  @override
  State<_VolunteerMatchSheet> createState() => _VolunteerMatchSheetState();
}

class _VolunteerMatchSheetState extends State<_VolunteerMatchSheet> {
  List<Map<String, dynamic>> _matches = [];
  bool _isLoading = true;
  final Set<String> _selectedIds = {};

  @override
  void initState() {
    super.initState();
    _loadMatches();
  }

  Future<void> _loadMatches() async {
    try {
      final matches = await widget.api.getMatchingVolunteers(widget.reportId);
      setState(() {
        _matches = List<Map<String, dynamic>>.from(matches);
        _isLoading = false;
      });
    } catch (e) {
      debugPrint('Match error: $e');
      setState(() => _isLoading = false);
    }
  }

  Future<void> _assignSelected() async {
    if (_selectedIds.isEmpty) return;
    try {
      await widget.api.assignVolunteers(widget.reportId, _selectedIds.toList());
      widget.onAssigned();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(const SnackBar(content: Text('Assignment failed')));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return DraggableScrollableSheet(
      initialChildSize: 0.7,
      maxChildSize: 0.9,
      minChildSize: 0.4,
      builder: (_, controller) => Container(
        decoration: BoxDecoration(
          color: theme.scaffoldBackgroundColor,
          borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
        ),
        child: Column(
          children: [
            // Handle
            Container(
              margin: const EdgeInsets.only(top: 8),
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                  color: theme.dividerColor,
                  borderRadius: BorderRadius.circular(2)),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  const Icon(Icons.people_alt, size: 22),
                  const SizedBox(width: 8),
                  Text('Matching Volunteers',
                      style: theme.textTheme.titleMedium
                          ?.copyWith(fontWeight: FontWeight.w600)),
                  const Spacer(),
                  FilledButton.icon(
                    onPressed: _selectedIds.isEmpty ? null : _assignSelected,
                    icon: const Icon(Icons.check, size: 16),
                    label: Text('Assign ${_selectedIds.length}'),
                  ),
                ],
              ),
            ),
            const Divider(height: 1),
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _matches.isEmpty
                      ? Center(
                          child: Column(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(Icons.person_search,
                                  size: 48,
                                  color: theme.colorScheme.outlineVariant),
                              const SizedBox(height: 12),
                              Text('No matching volunteers found',
                                  style: theme.textTheme.bodyLarge),
                              const SizedBox(height: 4),
                              Text('Volunteers will appear when they sign up',
                                  style: theme.textTheme.bodySmall),
                            ],
                          ),
                        )
                      : ListView.builder(
                          controller: controller,
                          itemCount: _matches.length,
                          itemBuilder: (context, index) {
                            final m = _matches[index];
                            final id = m['volunteer_id'] as String? ?? '';
                            final isSelected = _selectedIds.contains(id);
                            final score = (m['total_score'] as num?)
                                    ?.toStringAsFixed(1) ??
                                '0';

                            return CheckboxListTile(
                              value: isSelected,
                              onChanged: (v) => setState(() {
                                if (v == true) {
                                  _selectedIds.add(id);
                                } else {
                                  _selectedIds.remove(id);
                                }
                              }),
                              title: Text(
                                  m['volunteer_name'] as String? ?? 'Volunteer',
                                  style: const TextStyle(
                                      fontWeight: FontWeight.w600)),
                              subtitle: Text(
                                'Score: $score  |  '
                                'Skill: ${((m['skill_score'] as num?)?.toStringAsFixed(1) ?? '0')}  |  '
                                'Distance: ${((m['distance_score'] as num?)?.toStringAsFixed(1) ?? '0')}',
                              ),
                              secondary: CircleAvatar(
                                backgroundColor:
                                    theme.colorScheme.primaryContainer,
                                child: Text('${index + 1}',
                                    style: TextStyle(
                                      fontWeight: FontWeight.bold,
                                      color:
                                          theme.colorScheme.onPrimaryContainer,
                                    )),
                              ),
                            );
                          },
                        ),
            ),
          ],
        ),
      ),
    );
  }
}
