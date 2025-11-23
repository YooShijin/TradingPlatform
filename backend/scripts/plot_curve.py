import sys
import os
import pandas as pd
import matplotlib.pyplot as plt

# ----- Check CSV input -----
if len(sys.argv) < 2:
    print("Usage: python plot_load_test.py <path_to_metrics.csv>")
    sys.exit(1)

csv_path = sys.argv[1]

# Create output folder
OUTPUT_DIR = "plots_output"
os.makedirs(OUTPUT_DIR, exist_ok=True)

# Load CSV
df = pd.read_csv(csv_path)

# Extract columns
clients = df["clients"]
throughput = df["throughput"]
latency = df["avg_latency"]
cpu = df["cpu"]
disk = df["disk"]

plt.style.use("seaborn-v0_8")   # 🔥 Clean modern styling
marker_style = dict(marker='o', markersize=7, linewidth=2)

# -------- Plot: Clients vs Throughput --------
plt.figure(figsize=(9, 6))
plt.plot(clients, throughput, **marker_style)
plt.title("Clients vs Throughput", fontsize=16, fontweight='bold')
plt.xlabel("Number of Clients", fontsize=13)
plt.ylabel("Throughput (requests/sec)", fontsize=13)
plt.grid(True, linestyle="--", alpha=0.6)
plt.tight_layout()
plt.savefig(f"{OUTPUT_DIR}/clients_vs_throughput.png")
plt.close()

# -------- Plot: Clients vs Average Latency --------
plt.figure(figsize=(9, 6))
plt.plot(clients, latency, **marker_style)
plt.title("Clients vs Average Latency (Response Time)", fontsize=16, fontweight='bold')
plt.xlabel("Number of Clients", fontsize=13)
plt.ylabel("Latency (seconds)", fontsize=13)
plt.grid(True, linestyle="--", alpha=0.6)
plt.tight_layout()
plt.savefig(f"{OUTPUT_DIR}/clients_vs_latency.png")
plt.close()

# -------- Plot: Clients vs CPU Usage --------
plt.figure(figsize=(9, 6))
plt.plot(clients, cpu, **marker_style)
plt.title("Clients vs CPU Usage", fontsize=16, fontweight='bold')
plt.xlabel("Number of Clients", fontsize=13)
plt.ylabel("CPU Utilization (%)", fontsize=13)
plt.grid(True, linestyle="--", alpha=0.6)
plt.tight_layout()
plt.savefig(f"{OUTPUT_DIR}/clients_vs_cpu.png")
plt.close()

# -------- Plot: Clients vs Disk I/O --------
plt.figure(figsize=(9, 6))
plt.plot(clients, disk, **marker_style)
plt.title("Clients vs Disk I/O", fontsize=16, fontweight='bold')
plt.xlabel("Number of Clients", fontsize=13)
plt.ylabel("Disk (% util)", fontsize=13)
plt.grid(True, linestyle="--", alpha=0.6)
plt.tight_layout()
plt.savefig(f"{OUTPUT_DIR}/clients_vs_disk.png")
plt.close()

print(f"\n✨ Graphs generated inside folder: {OUTPUT_DIR}")
print("📌 Files:")
print(" - clients_vs_throughput.png")
print(" - clients_vs_latency.png")
print(" - clients_vs_cpu.png")
print(" - clients_vs_disk.png")
