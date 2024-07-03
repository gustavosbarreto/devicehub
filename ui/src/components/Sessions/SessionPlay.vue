<template>
  <div>
    <v-list-item
      @click="showDialog = !showDialog"
      v-bind="$attrs"
      :disabled="notHasAuthorization"
    >
      <div class="d-flex align-center">
        <v-icon class="mr-2"> mdi-play </v-icon>

        <v-list-item-title data-test="mdi-information-list-item">
          Play Session
        </v-list-item-title>
      </div>
    </v-list-item>

    <v-dialog
      :transition="false"
      v-model="showDialog"
      max-width="1024px"
      min-width="350px"
    >
      <v-card class="bg-v-theme-surface">
        <v-card-title
          class="text-h5 pa-3 bg-primary d-flex justify-space-between align-center"
        >
          Watch Session
          <v-btn
            variant="text"
            data-test="close-btn"
            icon="mdi-close"
            @click="showDialog = false"
          />
        </v-card-title>

        <v-card-item class="ma-0 pa-0 w-100">
          <div ref="terminal" />
        </v-card-item>

        <v-card-actions class="bg-v-theme-surface">
          <v-container>
            <v-row no-gutters>
              <v-col cols="2" sm="6" md="1">
                <div class="pt-4 ml-7">
                  <v-icon
                    v-if="!paused"
                    variant="text"
                    icon="mdi-pause-circle"
                    class="pl-0"
                    color="primary"
                    rounded
                    size="x-large"
                    data-test="pause-icon"
                    @click="pauseHandler"
                  />
                  <v-icon
                    v-else
                    variant="text"
                    icon="mdi-play-circle"
                    class="pl-0"
                    color="primary"
                    rounded
                    size="x-large"
                    data-test="play-icon"
                    @click="pauseHandler"
                  />
                </div>
              </v-col>

              <v-col cols="6" md="9">
                <div
                  :elevation="0"
                  class="pt-4 pl-9 mr-5 d-flex align-center"
                  tile
                >
                  <v-slider
                    v-model="currentTime"
                    class="ml-0"
                    min="0"
                    :max="totalLength"
                    :label="`${endTimerDisplay} - ${nowTimerDisplay} `"
                    hide-details
                    color="primary"
                    data-test="time-slider"
                    @update:model-value="changeSliderTime()"
                    @mousedown="(previousPause = paused), (paused = true)"
                    @mouseup="paused = previousPause"
                    @click="setSliderDiplayTime(currentTime)"
                  />
                </div>
              </v-col>

              <v-col cols="6" md="2" v-if="!props.isActive">
                <div :elevation="0">
                  <v-select
                    :items="speedList"
                    v-model="defaultSpeed"
                    hide-details
                    prepend-icon="mdi-speedometer"
                    data-test="speed-select"
                    variant="underlined"
                    color="primary"
                    @change="speedChange(defaultSpeed)"
                  />
                </div>
              </v-col>
              <v-col cols="6" md="2" v-else-if="!liveEnded">
                <h3 class="mt-4 red-live text-red">
                  <v-icon
                    size="24"
                    color="red"
                    icon="mdi-broadcast"
                    data-test="live-icon"
                  />
                  Live Viewing
                </h3>
              </v-col>
              <v-col cols="6" md="2" v-else @click="updatePlay()">
                <v-tooltip
                  location="bottom"
                  class="text-center"
                >
                  <template v-slot:activator="{ props }">
                    <h3 v-bind="props" class="mt-4 text-grey">
                      <v-icon
                        size="24"
                        icon="mdi-broadcast"
                        data-test="live-stoped-icon"
                      />
                      Live Stopped
                    </h3>
                  </template>
                  <span>The live viewing stops because the terminal isn't capturing anything new,
                    please click the play button or the Live Stopped button to refresh the Session.</span>
                </v-tooltip>
              </v-col>
            </v-row>
          </v-container>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  nextTick,
  onUpdated,
  ref,
  watch,
} from "vue";
import { Terminal } from "xterm";
import "xterm/css/xterm.css";
import { FitAddon } from "xterm-addon-fit";
import moment from "moment";
import axios from "axios";
import { useStore } from "../../store";
import { INotificationsError } from "../../interfaces/INotifications";
import handleError from "../../utils/handleError";
import { ITerminalFrames, ITerminalLog } from "../../interfaces/ITerminal";

type Timer = ReturnType<typeof setTimeout>;

const props = defineProps({
  uid: {
    type: String,
    required: true,
  },
  recorded: {
    type: Boolean,
    required: true,
  },
  notHasAuthorization: {
    type: Boolean,
    default: false,
  },
  isActive: {
    type: Boolean,
    required: true,
  },
});
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const terminal = ref<HTMLElement>({} as HTMLElement);
const currentTime = ref(0);
const totalLength = ref(0);
const endTimerDisplay = ref<string | number>(0);
const getTimerNow = ref<string | number>(0);
const paused = ref(false);
const previousPause = ref(false);
const sliderChange = ref(false);
const speedList = ref([0.5, 1, 1.5, 2, 4]);
const logs = ref<Array<ITerminalLog>>([]);
const frames = ref<Array<ITerminalFrames>>([]);
const defaultSpeed = ref(1);
const transition = ref(false);
const xterm = ref<Terminal>({} as Terminal);
const fitAddon = ref<FitAddon>({} as FitAddon);
const iterativeTimer = ref<Timer>();
const iterativePrinting = ref<Timer>();
const liveEnded = ref(false);
const store = useStore();
const length = computed(() => logs.value.length);
const nowTimerDisplay = computed(() => getTimerNow.value);

const getSliderIntervalLength = (timeMs: number | null) => {
  let interval: number;
  if (!timeMs && logs.value.length > 0) {
    // not params, will return metrics to max timelength
    const max = new Date(logs.value[length.value - 1].time);
    const min = new Date(logs.value[0].time);
    interval = +max - +min;
  } else {
    // it will format to the time argument passed
    interval = timeMs || 0;
  }

  return interval;
};

const setSliderDiplayTime = async (timeMs: number | null) => {
  const interval = getSliderIntervalLength(timeMs);
  const duration = moment.duration(interval, "milliseconds");

  // format according to how long
  let hoursFormat;
  if (duration.asHours() > 1) hoursFormat = "h";
  else hoursFormat = "";

  const displayTime = moment
    .utc(duration.asMilliseconds())
    .format(`${hoursFormat ? `${hoursFormat}:` : ""}mm:ss`);
  if (timeMs) {
    endTimerDisplay.value = displayTime;
  } else {
    getTimerNow.value = displayTime;
  }
};

const createFrames = () => {
  // create cumulative frames for the exibition in slider
  let time = 0;
  let message = "";
  const arrFrames = [
    {
      incMessage: (message += logs.value[0].message),
      incTime: time,
    },
  ];

  for (let i = 1; i < logs.value.length; i += 1) {
    const future = new Date(logs.value[i].time);
    const now = new Date(logs.value[i - 1].time);
    const interval = moment
      .duration(+future - +now, "milliseconds")
      .asMilliseconds();
    time += interval;
    message += logs.value[i].message;
    arrFrames.push({
      incMessage: message,
      incTime: time,
    });
  }
  return arrFrames;
};

const timer = () => {
  // Increments the slider
  if (!paused.value) {
    if (currentTime.value >= totalLength.value) {
      paused.value = true;
      return;
    }
    currentTime.value += 100;
    setSliderDiplayTime(currentTime.value);
  }
  iterativeTimer.value = setTimeout(
    timer.bind(null),
    100 * (1 / defaultSpeed.value),
  );
};

const searchClosestFrame = (givenTime: number, frames: Array<ITerminalFrames>) => {
  // applies a binary search to find nearest frame
  let between: number;
  let lowerBound = 0;
  let higherBound = frames.length - 1;
  let nextTimeSetPrint;

  for (; higherBound - lowerBound > 1;) {
    // progressive increment search
    between = Math.floor((lowerBound + higherBound) / 2);
    if (frames[between].incTime < givenTime) {
      lowerBound = between;
      nextTimeSetPrint = givenTime - frames[between].incTime;
    } else {
      higherBound = between;
      nextTimeSetPrint = frames[between].incTime - givenTime;
    }
  }
  return {
    message: frames[lowerBound].incMessage,
    index: lowerBound,
    waitForPrint: nextTimeSetPrint,
  };
};

const print = (i: number, logsArray: Array<ITerminalLog>) => {
  // Writes iteratevely on xterm as time progresses
  sliderChange.value = false;
  if (!paused.value) {
    xterm.value.write(`${logsArray[i].message}`);
    if (i === logsArray.length - 1) return;
    const nowTimerDisplay = new Date(logsArray[i].time);
    const future = new Date(logsArray[i + 1].time);
    const interval = +future - +nowTimerDisplay;
    iterativePrinting.value = setTimeout(
      print.bind(null, i + 1, logsArray),
      interval * (1 / defaultSpeed.value),
    );
  }
};

const clear = () => {
  // Ensure to clear functions for syncronism
  clearInterval(iterativePrinting.value);
  clearInterval(iterativeTimer.value);
};

const xtermSyncFrame = (givenTime: number) => {
  if (xterm.value) {
    xterm.value.write("\u001Bc"); // clean screen
    const frame = searchClosestFrame(givenTime, frames.value);
    clear();
    xterm.value.write(frame.message); // write frame on xterm
    iterativeTimer.value = setTimeout(timer.bind(null), 1);
    iterativePrinting.value = setTimeout(
      print.bind(null, frame.index + 1, logs.value),
      frame.waitForPrint * (1 / defaultSpeed.value),
    );
  }
};

const speedChange = (speed: number) => {
  defaultSpeed.value = speed;
  xtermSyncFrame(currentTime.value);
};

const changeSliderTime = () => {
  sliderChange.value = true;
  xtermSyncFrame(currentTime.value);
};

onUpdated(() => {
  if (showDialog.value) {
    setSliderDiplayTime(currentTime.value);
  }
});

const openPlay = async () => {
  if (props.recorded) {
    await store.dispatch("sessions/getLogSession", props.uid);
    logs.value = store.getters["sessions/get"];
    totalLength.value = getSliderIntervalLength(null);
    setSliderDiplayTime(null);
    setSliderDiplayTime(currentTime.value);

    frames.value = createFrames();

    xterm.value = new Terminal({
      cursorBlink: true,
      fontFamily: "monospace",
      theme: {
        background: "#0f1526",
      },
    });

    fitAddon.value = new FitAddon();
    xterm.value.loadAddon(fitAddon.value);

    if (xterm.value.element) {
      xterm.value.reset();
    }
  }
};

const updatePlay = async () => {
  emit("update");
  if (props.isActive && props.recorded) {
    try {
      await store.dispatch("sessions/updateLogSession", props.uid);
      logs.value = store.getters["sessions/get"];
      totalLength.value = getSliderIntervalLength(null);
      setSliderDiplayTime(null);
      setSliderDiplayTime(currentTime.value);

      frames.value = createFrames();

      xterm.value.loadAddon(fitAddon.value);
      xtermSyncFrame(currentTime.value);

      liveEnded.value = false;

      if (xterm.value.element) {
        xterm.value.reset();
      }
    } catch (error) {
      if (axios.isAxiosError(error)) {
        return;
      }
      liveEnded.value = true;
    }
  }
};

const pauseHandler = () => {
  paused.value = !paused.value;

  if (props.isActive) {
    setTimeout(() => {
      updatePlay();
    }, 3000);
  }
};

watch(
  () => currentTime.value >= (totalLength.value - 5),
  (newValue) => {
    if (newValue) {
      updatePlay();
    }
  },
);

const connect = async () => {
  if (!xterm.value.element) {
    xterm.value.open(terminal.value);
    fitAddon.value.fit();
    xterm.value.focus();
    print(0, logs.value);
    timer();
  }
};

const displayDialog = async () => {
  try {
    await openPlay();

    await nextTick().then(() => {
      connect();
    });
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorLoading",
      INotificationsError.sessionPlay,
    );
    handleError(error);
  }
};

const close = async () => {
  transition.value = true;
  if (xterm.value) {
    xterm.value.reset();
    xterm.value.element?.remove();
  }
  clear();
  currentTime.value = 0;
  paused.value = false;
  defaultSpeed.value = 1;
  liveEnded.value = false;
  emit("update");
};

watch(showDialog, (value) => {
  if (!value) {
    close();
    showDialog.value = false;
  } else {
    displayDialog();
  }
});
</script>

<style scoped>
.red-live {
  filter: drop-shadow(0px 0px 10px rgba(255, 0, 0, 0.444))
}</style>
