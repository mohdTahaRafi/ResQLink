import 'package:flutter/material.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:go_router/go_router.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

/// Nagar Nigam "Command Mode" — Real-time heatmap of city reports.
class CommandDashboard extends StatefulWidget {
  const CommandDashboard({super.key});

  @override
  State<CommandDashboard> createState() => _CommandDashboardState();
}

class _CommandDashboardState extends State<CommandDashboard> {
  final Set<Circle> _heatmapCircles = {};
  bool _isLoading = true;

  // Lucknow center coordinates
  static const _lucknowCenter = LatLng(26.8467, 80.9462);

  // Simulated heatmap data — will be replaced by API calls
  final _mockHeatData = [
    {'lat': 26.8600, 'lng': 80.9400, 'intensity': 8.5, 'category': 'water'},
    {'lat': 26.8470, 'lng': 80.9550, 'intensity': 6.2, 'category': 'road'},
    {
      'lat': 26.8350,
      'lng': 80.9300,
      'intensity': 9.1,
      'category': 'sanitation'
    },
    {
      'lat': 26.8550,
      'lng': 80.9700,
      'intensity': 4.0,
      'category': 'electricity'
    },
    {'lat': 26.8700, 'lng': 80.9200, 'intensity': 7.3, 'category': 'health'},
  ];

  @override
  void initState() {
    super.initState();
    _loadHeatmapData();
  }

  Future<void> _loadHeatmapData() async {
    // TODO: Fetch from API /api/v1/dashboard/nagar_nigam
    await Future.delayed(const Duration(milliseconds: 500));

    final circles = <Circle>{};
    for (int i = 0; i < _mockHeatData.length; i++) {
      final point = _mockHeatData[i];
      final intensity = (point['intensity'] as double);
      circles.add(Circle(
        circleId: CircleId('heat_$i'),
        center: LatLng(point['lat'] as double, point['lng'] as double),
        radius: intensity * 150,
        fillColor:
            _categoryColor(point['category'] as String).withValues(alpha: 0.3),
        strokeColor: _categoryColor(point['category'] as String),
        strokeWidth: 2,
      ));
    }

    setState(() {
      _heatmapCircles.addAll(circles);
      _isLoading = false;
    });
  }

  Color _categoryColor(String category) {
    switch (category) {
      case 'water':
        return Colors.blue;
      case 'road':
        return Colors.orange;
      case 'sanitation':
        return Colors.red;
      case 'electricity':
        return Colors.amber;
      case 'health':
        return Colors.green;
      default:
        return Colors.purple;
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Command Mode'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () {
              setState(() => _isLoading = true);
              _loadHeatmapData();
            },
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await FirebaseAuth.instance.signOut();
              if (!context.mounted) return;
              context.go('/login');
            },
          ),
        ],
      ),
      body: Column(
        children: [
          // Legend bar
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            color: theme.colorScheme.surfaceContainerLow,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _LegendDot(color: Colors.blue, label: 'Water'),
                _LegendDot(color: Colors.orange, label: 'Road'),
                _LegendDot(color: Colors.red, label: 'Sanitation'),
                _LegendDot(color: Colors.amber, label: 'Electric'),
                _LegendDot(color: Colors.green, label: 'Health'),
              ],
            ),
          ),

          // Map
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : GoogleMap(
                    initialCameraPosition: const CameraPosition(
                      target: _lucknowCenter,
                      zoom: 13,
                    ),
                    circles: _heatmapCircles,
                    myLocationEnabled: true,
                    myLocationButtonEnabled: true,
                    mapToolbarEnabled: false,
                    zoomControlsEnabled: false,
                  ),
          ),

          // Quick stats footer
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: theme.colorScheme.surface,
              boxShadow: [
                BoxShadow(
                    color: Colors.black12,
                    blurRadius: 8,
                    offset: const Offset(0, -2))
              ],
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _QuickStat(
                    label: 'Active',
                    value: '${_mockHeatData.length}',
                    icon: Icons.warning_amber),
                _QuickStat(
                    label: 'Critical',
                    value:
                        '${_mockHeatData.where((d) => (d['intensity'] as double) > 7).length}',
                    icon: Icons.error),
                _QuickStat(
                    label: 'Wards', value: '5', icon: Icons.location_city),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _LegendDot extends StatelessWidget {
  final Color color;
  final String label;

  const _LegendDot({required this.color, required this.label});

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

class _QuickStat extends StatelessWidget {
  final String label;
  final String value;
  final IconData icon;

  const _QuickStat(
      {required this.label, required this.value, required this.icon});

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 22, color: Theme.of(context).colorScheme.primary),
        const SizedBox(height: 4),
        Text(value,
            style: Theme.of(context)
                .textTheme
                .titleMedium
                ?.copyWith(fontWeight: FontWeight.bold)),
        Text(label, style: Theme.of(context).textTheme.labelSmall),
      ],
    );
  }
}
