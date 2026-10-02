<template>
  <content-area style="padding: 0 !important">
    <eq-window title="Export client files">
      <div class="ui-toolbar">
        <div class="ui-stat-line mr-auto">Download the current database as client files.</div>
        <b-button @click="downloadSpells" size="sm" variant="primary">
          <i class="fa fa-cloud-download"></i>
          spells_us.txt
        </b-button>
        <b-button @click="downloadDbStr" size="sm" variant="primary">
          <i class="fa fa-cloud-download"></i>
          dbstr_us.txt
        </b-button>
      </div>
      <div class="ui-stat-line" v-if="successMessage">{{ successMessage }}</div>
      <loader-fake-progress v-if="loading" class="mt-2"/>
    </eq-window>

    <eq-window title="Import client files">
      <div class="ui-problem mb-3">
        Dropped files overwrite matching database values immediately.
      </div>
      <vue-dropzone
        style="min-height: 360px"
        v-on:vdropzone-success="success"
        v-on:vdropzone-queue-complete="queueComplete"
        v-on:vdropzone-processing="processing"
        ref="myVueDropzone"
        id="dropzone"
        :options="dropzoneOptions"
      />
    </eq-window>
  </content-area>
</template>

<script>
import querystring        from 'querystring'
import vue2Dropzone       from 'vue2-dropzone'
import 'vue2-dropzone/dist/vue2Dropzone.min.css'
import {SpireApi}         from "../../app/api/spire-api";
import EqWindowSimple     from "../../components/eq-ui/EQWindowSimple";
import EqWindow           from "../../components/eq-ui/EQWindow";
import LoaderFakeProgress from "../../components/LoaderFakeProgress";
import util               from "util";
import ContentArea        from "../../components/layout/ContentArea";

export default {
  name: "ClientFiles.vue",
  components: {
    ContentArea,
    LoaderFakeProgress,
    EqWindow,
    EqWindowSimple,
    vueDropzone: vue2Dropzone
  },
  data: function () {
    return {
      successMessage: "",
      loading: false,

      dropzoneOptions: {
        url: util.format(
          "%s/api/v1/client-files/import/file?%s",
          SpireApi.getBasePath(),
          querystring.stringify(SpireApi.getAccessTokenQueryString())
        ),
        resizeWidth: 50,
        resizeHeight: 50,
        thumbnailWidth: 150,
        thumbnailHeight: 150,
        thumbnailMethod: "crop",
        // maxFilesize: 0.5,
        addRemoveLinks: true,
      }
    }
  },
  methods: {

    downloadSpells() {
      window.open(
        util.format(
          "%s/api/v1/client-files/export/spells?%s",
          SpireApi.getBasePath(),
          querystring.stringify(SpireApi.getAccessTokenQueryString())
        ),
        '_blank'
      );
    },
    downloadDbStr() {
      window.open(
        util.format(
          "%s/api/v1/client-files/export/dbstr?%s",
          SpireApi.getBasePath(),
          querystring.stringify(SpireApi.getAccessTokenQueryString())
        ),
        '_blank'
      );
    },

    success(file, response) {
      this.loading = false
      console.log("file", file)
      console.log("response", response)

      this.successMessage = response

      setTimeout(() => {
        this.$refs.myVueDropzone.removeAllFiles()
      }, 1000)

      setTimeout(() => {
        this.successMessage = ""
      }, 6000)
    },
    queueComplete() {
      console.log("complete")
    },
    processing(event) {
      this.loading = true
      console.log("dropped")
      console.log(event)
    }
  }
}
</script>

<style>
.vue-dropzone:hover, .vue-dropzone {
  background-color: rgba(0, 0, 0, 0.6);
}

.dz-message {
  font-size: 14px;
  background-color: #161a25;
  color: yellow;
}

.dz-message:hover {
  color: yellow;
}

.dz-message {
  height: 80vh;
}

</style>
