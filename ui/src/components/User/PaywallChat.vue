<template>
  <v-dialog
    v-model="dialog"
    transition="dialog-bottom-transition"
    max-width="600"
    @click:outside="close"
    @keydown.esc="close"
  >
    <v-card color="background" data-test="card-dialog">
      <v-container>
        <v-row data-test="icon-chat">
          <v-col class="d-flex justify-center align-center">
            <div class="circle-one shadow d-flex justify-center align-center">
              <div class="circle-two shadow d-flex justify-center align-center">
                <v-icon color="success" class="green-inner-shadow" size="108">
                  mdi-chat-question
                </v-icon>
              </div>
            </div>
          </v-col>
        </v-row>
        <v-row>
          <v-col class="pb-0">
            <h1 class="d-flex justify-center align-center text-center" data-test="upgrade-heading">
              Upgrade to have access to chat support!
            </h1>
          </v-col>
        </v-row>
        <v-row class="px-6">
          <v-col class="pt-1 pb-6">
            <p class="d-flex justify-center align-center text-grey text-center" data-test="upgrade-description">
              Get real-time assistance from our team with priority responses.
              Skip the hassle of documentation — upgrade now and unlock direct chat support.
            </p>
            <p class="align-center text-grey text-center" data-test="upgrade-description2">
              However, you can still use
              <a href="https://docs.shellhub.io/" target="_blank" rel="noopener noreferrer" data-test="link-anchor">our Documentation</a>
              to find answers to your questions and troubleshoot issues on your own.
            </p>
          </v-col>
        </v-row>
        <v-card-actions data-test="card-actions">
          <v-btn @click="close()" class="mt-4" variant="text" data-test="close-btn">
            Close
          </v-btn>
          <v-spacer />
          <v-btn
            color="primary"
            variant="flat"
            href="www.shellhub.io/pricing"
            target="_blank"
            rel="noreferrer noopener"
            data-test="upgrade-btn">
            Upgrade
          </v-btn>
        </v-card-actions>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
const dialog = defineModel({ default: false });

const close = () => {
  dialog.value = false;
};

defineExpose({ dialog });
</script>

<style scoped>
.green-inner-shadow {
  filter: drop-shadow(0px 0px 20px rgba(43, 255, 10, 0.45));
}

.shadow {
  box-shadow: rgba(0, 0, 0, 0.35) 0px 15px 10px 0px;
  border-radius: 50%;
}

.circle-one {
  height: 15.625rem;
  width: 15.625rem;
  background: linear-gradient(180deg, rgba(0,0,0,0) 40%, rgba(76,175,80,0.1) 60%, rgba(76,175,80,0.75) 120%);
}

.circle-two {
  height: 12.5rem;
  width: 12.5rem;
  background: linear-gradient(180deg, rgba(0,0,0,0) 60%, rgba(76,175,80,0.1) 75%, rgba(76,175,80,0.75) 120%);
}
</style>
