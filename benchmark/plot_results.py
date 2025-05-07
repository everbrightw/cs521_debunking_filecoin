import pandas as pd
import matplotlib.pyplot as plt
import os
import platform

df = pd.read_csv("results/benchmark_log.csv")

fig, ax = plt.subplots(figsize=(8, 6))
x_labels = []
upload_times = []
download_times = []

for idx, row in df.iterrows():
    label = f"{os.path.basename(row['File'])}\n{row['Chunker']}"
    x_labels.append(label)
    upload_times.append(row["UploadTime"])
    download_times.append(row["DownloadTime"])

x = range(len(x_labels))
bar_width = 0.35

ax.bar([i - bar_width/2 for i in x], upload_times, width=bar_width, label="Upload Time")
ax.bar([i + bar_width/2 for i in x], download_times, width=bar_width, label="Download Time", hatch='//')

ax.set_xticks(x)
ax.set_xticklabels(x_labels, rotation=30, ha="right")
ax.set_ylabel("Time (s)")
ax.set_title("IPFS Benchmark: Upload vs Download Time")
ax.legend()
plt.tight_layout()

plot_path = "plots/benchmark_plot.png"
plt.savefig(plot_path)
print(f"✅ Plot saved to {plot_path}")

if platform.system() == "Darwin":
    os.system(f"open {plot_path}")
elif platform.system() == "Linux":
    os.system(f"xdg-open {plot_path}")

plt.show()