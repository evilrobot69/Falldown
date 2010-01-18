package com.ariwilson.seismo;

import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.LinkedList;
import java.util.concurrent.LinkedBlockingQueue;

import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Paint;
import android.hardware.SensorManager;
import android.view.SurfaceHolder;
import android.widget.Toast;

public class SeismoViewThread extends Thread {
  public SeismoViewThread(Context ctx, SurfaceHolder holder, boolean filter,
                          int axis, int period) {
    holder_ = holder;
    setFilter(filter);
    setAxis(axis);
    db_ = SeismoDbAdapter.getAdapter();
    ctx_ = ctx;
    period_ = period;
    profiler_ = new Profiler();  // TODO(ariw): REMOVE
  }

  @Override
  public void run() {
    profiler_.start();  // TODO(ariw): REMOVE
    while (running_) {
      profiler_.record("overhead");  // TODO(ariw): REMOVE
      // Retrieve measurements from queue.
      LinkedList<ArrayList<Float>> accelerations =
          new LinkedList<ArrayList<Float>>();
      history_queue_.drainTo(accelerations);
      synchronized (history_) {
        for (ArrayList<Float> acceleration : accelerations) {
          history_.add(acceleration);
          while (acceleration.get(0) - history_.get(0).get(0) >
                 SECONDS_TO_SAVE * 1000) {
            history_.remove(0);
          }
          while (acceleration.get(0) - history_.get(start_).get(0) >
                 SECONDS_TO_DISPLAY * 1000) {
            ++start_;
          }
        }
        profiler_.record("history queue draining");  // TODO(ariw): REMOVE
        synchronized (holder_) {
          Canvas canvas = holder_.lockCanvas();
          canvas.drawARGB(255, 255, 255, 255);
  
          Paint scale_paint = new Paint();
          scale_paint.setARGB(255, 137, 137, 137);
          scale_paint.setStrokeWidth(canvas_width_ / 300f);
          scale_paint.setAntiAlias(true);
          float text_size = canvas_width_ / 35f;
          scale_paint.setTextSize(text_size);
          
          // Draw g scale.
          scale_paint.setTextAlign(Paint.Align.CENTER);
          for (int i = -MAX_G + 1; i <= MAX_G - 1; ++i) {
            float x = canvas_width_ / 2 * (1 + (float)i / MAX_G);
            canvas.drawLine(x, 0, x, canvas_height_ / 20, scale_paint);
            canvas.drawText(Integer.toString(i) + "g", x,
                            canvas_height_ / 20 + 1.2f * text_size,
                            scale_paint);
          }
          profiler_.record("drawing background plus g scale");  // TODO(ariw): REMOVE
  
          // Draw time scale in seconds.
          // Don't want to determine scale if no values written yet.
          float end_time = -1, start_time = -1;
          if (history_.size() > 0) {
            end_time = history_.get(history_.size() - 1).get(0) / 1000;
            start_time = end_time - SECONDS_TO_DISPLAY;
            scale_paint.setTextAlign(Paint.Align.LEFT);
            for (int s = (int) Math.floor(end_time);
                 s >= Math.max(Math.floor(start_time), 0);
                 --s) {
              float y = canvas_height_ * (s - start_time) / SECONDS_TO_DISPLAY;
              canvas.drawLine(0, y, canvas_width_ / 20, y, scale_paint);
              canvas.drawText(Integer.toString(s) + "s",
                              canvas_width_ / 20 + 0.2f * text_size,
                              y + 0.5f * text_size, scale_paint);
            }
          }
          profiler_.record("drawing time scale");  // TODO(ariw): REMOVE

          drawFPS(canvas);  // TODO(ariw): REMOVE
          profiler_.record("drawing FPS");  // TODO(ariw): REMOVE

          // Draw line.
          float[] pts = new float[(history_.size() - start_) * 4];
          for (int i = start_ + 1; i < history_.size(); ++i) {
            ArrayList<Float> history1 = history_.get(i - 1),
                             history2 = history_.get(i);
            int j = i - start_ - 1;
            pts[j * 4] = canvas_width_ / 2 *
                         (1 + history1.get(axis_ + 1) / MAX_ACCELERATION);
            pts[j * 4 + 1] = canvas_height_ *
                             (history1.get(0) / 1000 - start_time) /
                             SECONDS_TO_DISPLAY;
            pts[j * 4 + 2] = canvas_width_ / 2 *
                             (1 + history2.get(axis_ + 1) / MAX_ACCELERATION);
            pts[j * 4 + 3] = canvas_height_ *
                             (history2.get(0) / 1000 - start_time) /
                             SECONDS_TO_DISPLAY;
          }
          Paint line_paint = new Paint();
          line_paint.setARGB(255, 0, 0, 0);
          line_paint.setStrokeWidth(canvas_width_ / 300f);
          line_paint.setAntiAlias(false);
          canvas.drawLines(pts, line_paint);
          profiler_.record("drawing awesome line");  // TODO(ariw): REMOVE
          holder_.unlockCanvasAndPost(canvas);
        }
      }
      try {
        Thread.sleep(period_);
      } catch (Exception e) {
        // Ignore.
      }
    }
  }

  public void update(float x, float y, float z) {
    ArrayList<Float> acceleration = new ArrayList<Float>(3);
    acceleration.add((float)(new Date().getTime() - start_time_));
    if (filter_) {
      filter_acceleration_[0] = x * FILTERING_FACTOR +
                         filter_acceleration_[0] * (1.0f - FILTERING_FACTOR);
      acceleration.add(x - filter_acceleration_[0]);
      filter_acceleration_[1] = y * FILTERING_FACTOR +
                         filter_acceleration_[1] * (1.0f - FILTERING_FACTOR);
      acceleration.add(y - filter_acceleration_[1]);
      filter_acceleration_[2] = z * FILTERING_FACTOR +
                         filter_acceleration_[2] * (1.0f - FILTERING_FACTOR);
      acceleration.add(z - filter_acceleration_[2]);
    } else {
      acceleration.add(x);
      acceleration.add(y);
      acceleration.add(z);
    }
    try {
      history_queue_.put(acceleration);
    } catch (Exception e) {
      // Do nothing.
    }
  }

  public void setSurfaceSize(int canvas_width, int canvas_height) {
    synchronized (holder_) {
      canvas_width_ = canvas_width;
      canvas_height_ = canvas_height;
      start_time_ = new Date().getTime();
      start_ = 0;
    }
  }

  public void setRunning(boolean running) {
    running_ = running;
  }

  public void setPaused(boolean paused) {
    if (paused) {
      start_paused_time_ = new Date().getTime();
    } else {
      synchronized (history_) {
        start_time_ += new Date().getTime() - start_paused_time_;
      }
    }
  }

  public void setFilter(boolean filter) {
    filter_ = filter;
  }
  
  public void setAxis(int axis) {
    axis_ = axis;
  }

  public void save() {
    SimpleDateFormat date_format = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
    Date date = new Date();
    String name = date_format.format(date);

    db_.open(ctx_);
    synchronized (history_) {
      if (db_.createGraph(name, history_) >= 0) {
        Toast.makeText(ctx_, "Saved graph as " + name + ".", Toast.LENGTH_LONG)
            .show();
      } else {
        Toast.makeText(ctx_, "Failed to save graph. Please try again.",
                       Toast.LENGTH_LONG).show();
      }
    }
    db_.close();
  }

  // TODO(ariw): REMOVE
  // FPS-related stuff.
  private long cur_second;
  private long last_second = -1;
  private int frames = 0;
  private int old_frames = 0;
  private void drawFPS(Canvas canvas) {
    cur_second = new Date().getTime() / 1000;
    if (cur_second != last_second) {
      if (cur_second % 15 == 0) {
        profiler_.print();
      }
      last_second = cur_second;
      old_frames = frames;
      frames = 0;
    }
    Paint text_paint = new Paint();
    text_paint.setARGB(255, 0, 0, 0);
    text_paint.setTextAlign(Paint.Align.RIGHT);
    canvas.drawText(Integer.toString(old_frames), canvas_width_, 100, text_paint);
    ++frames;
  }

  // Random constants.
  private static final int MAX_G = 3;
  private static final float MAX_ACCELERATION = MAX_G *
                             SensorManager.GRAVITY_EARTH;
  private static final float FILTERING_FACTOR = 0.1f;
  private static final int SECONDS_TO_SAVE = 60;
  private static final int SECONDS_TO_DISPLAY = 10;

  // Important preferences and history.
  // TODO(ariw): Worst data structure choice ever.
  private ArrayList<ArrayList<Float>> history_ =
      new ArrayList<ArrayList<Float>>();
  private LinkedBlockingQueue<ArrayList<Float>> history_queue_ =
      new LinkedBlockingQueue<ArrayList<Float>>();
  private int start_ = 0;
  private float[] filter_acceleration_ = new float[3];
  private long start_time_ = new Date().getTime();
  private long start_paused_time_ = start_time_;
  private int canvas_height_ = 1;
  private int canvas_width_ = 1;
  private boolean running_ = false;
  private boolean filter_ = true;
  private int axis_ = 2;
  private SeismoDbAdapter db_;
  private SurfaceHolder holder_;
  private Context ctx_;
  private int period_;
  private Profiler profiler_; // TODO(ariw): REMOVE
}
